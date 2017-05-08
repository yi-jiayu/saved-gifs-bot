package saved_gifs_bot

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/search"
)

// app engine search indexes
const (
	gifsIndex = "Gifs"
)

// app engine datastore kinds
const (
	packKind         = "Pack"
	subscriptionKind = "Subscription"
)

var (
	collapseWhitespaceRegex = regexp.MustCompile(`\s+`)
	packNameRegex           = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
)

// exported errors
var (
	ErrInvalidName = errors.New("invalid name")
	ErrNotAllowed  = errors.New("not allowed")
	ErrNotFound    = errors.New("pack not found")
)

type Gif struct {
	Pack     search.Atom
	FileID   search.Atom
	Keywords string
}

type Pack struct {
	Name    string
	Creator int
}

type Subscription struct {
	User int
	Pack string
}

// Returns true if pack was created, false if a pack with the same name already exists.
func NewPack(ctx context.Context, name string, creator int) (bool, error) {
	// validate pack name
	if !packNameRegex.MatchString(name) {
		return false, ErrInvalidName
	}

	// check if pack name is already taken
	var pack Pack
	key := datastore.NewKey(ctx, packKind, name, 0, nil)
	err := datastore.Get(ctx, key, &pack)
	if err != nil {
		if err != datastore.ErrNoSuchEntity {
			return false, err
		}
	} else {
		return false, nil
	}

	pack = Pack{
		Name:    name,
		Creator: creator,
	}

	_, err = datastore.Put(ctx, key, &pack)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Returns a slice of packs which were created by creator
func MyPacks(ctx context.Context, creator int) ([]Pack, error) {
	q := datastore.NewQuery(packKind).Filter("Creator =", creator)

	var packs []Pack
	_, err := q.GetAll(ctx, &packs)
	if err != nil {
		return nil, err
	}

	return packs, nil
}

// Retrieves information about a specific gif pack
func GetPack(ctx context.Context, name string) (Pack, error) {
	// validate pack name
	if !packNameRegex.MatchString(name) {
		return Pack{}, ErrInvalidName
	}

	key := datastore.NewKey(ctx, packKind, name, 0, nil)
	var pack Pack
	err := datastore.Get(ctx, key, &pack)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return Pack{}, ErrNotFound
		} else {
			return Pack{}, err
		}
	}

	return pack, nil
}

// Returns true if user was successfully subscribed to pack, false if user was already subscribed to pack.
// Returns ErrNotFound if pack does not exist.
func Subscribe(ctx context.Context, pack string, user int) (bool, error) {
	// validate pack name
	if !packNameRegex.MatchString(pack) {
		return false, ErrInvalidName
	}

	// check if pack exists
	var p Pack
	key := datastore.NewKey(ctx, packKind, pack, 0, nil)
	err := datastore.Get(ctx, key, &p)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return false, ErrNotFound
		} else {
			return false, err
		}
	}

	// check if user is already subscribed
	q := datastore.NewQuery(subscriptionKind).Filter("User =", user).Filter("Pack =", pack)

	var subscriptions []Subscription
	_, err = q.GetAll(ctx, &subscriptions)
	if err != nil {
		return false, err
	}

	if len(subscriptions) > 0 {
		return false, nil
	}

	subscription := Subscription{
		User: user,
		Pack: pack,
	}

	key = datastore.NewIncompleteKey(ctx, subscriptionKind, nil)
	_, err = datastore.Put(ctx, key, &subscription)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Returns true if user was successfully unsubscribed from pack, false if user was not subscribed to pack.
// err will be ErrInvalidName if pack is not a valid pack name
func Unsubscribe(ctx context.Context, pack string, user int) (bool, error) {
	// validate pack name
	if !packNameRegex.MatchString(pack) {
		return false, ErrInvalidName
	}

	// check if user is already subscribed
	q := datastore.NewQuery(subscriptionKind).Filter("User =", user).Filter("Pack =", pack)

	var keys []*datastore.Key
	var subscriptions []Subscription
	keys, err := q.GetAll(ctx, &subscriptions)
	if err != nil {
		return false, err
	}

	if len(keys) == 0 {
		return false, nil
	}

	// just delete all keys even though we only expect one to match
	for _, key := range keys {
		err := datastore.Delete(ctx, key)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

func MySubscriptions(ctx context.Context, user int) ([]Subscription, error) {
	q := datastore.NewQuery(subscriptionKind).Filter("User =", user)

	var subscriptions []Subscription
	_, err := q.GetAll(ctx, &subscriptions)
	if err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func NewGif(ctx context.Context, pack string, user int, gif Gif) error {
	// todo: return additional information about whether a gif is already in a pack

	// validate pack name
	if !packNameRegex.MatchString(pack) {
		return ErrInvalidName
	}

	// check that user is the creator of pack
	var p Pack
	key := datastore.NewKey(ctx, packKind, pack, 0, nil)
	err := datastore.Get(ctx, key, &p)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return ErrNotFound
		} else {
			return err
		}
	}

	if user != p.Creator {
		return ErrNotAllowed
	}

	index, err := search.Open(gifsIndex)
	if err != nil {
		return err
	}

	_, err = index.Put(ctx, "", &gif)
	if err != nil {
		return err
	}

	return nil
}

func DeleteGif(ctx context.Context, pack string, user int, fileId string) (int, error) {
	// validate pack name
	if !packNameRegex.MatchString(pack) {
		return 0, ErrInvalidName
	}

	// check that user is the creator of pack
	var p Pack
	key := datastore.NewKey(ctx, packKind, pack, 0, nil)
	err := datastore.Get(ctx, key, &p)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return 0, ErrNotFound
		} else {
			return 0, err
		}
	}

	if user != p.Creator {
		return 0, ErrNotAllowed
	}

	// get all instances of this gif in the pack
	index, err := search.Open(gifsIndex)
	if err != nil {
		return 0, err
	}

	deleted := 0
	q := fmt.Sprintf("Pack = %s AND FileID = %s", pack, fileId)
	for t := index.Search(ctx, q, nil); ; {
		var gif Gif
		id, err := t.Next(&gif)
		if err != nil {
			if err == search.Done {
				break
			} else {
				return deleted, err
			}
		}

		// try to delete all gifs which have the same fileid
		err = index.Delete(ctx, id)
		if err != nil {
			return deleted, err
		}
		deleted++
	}

	return deleted, nil
}

// query is a string with the format
//   <query> ::= <pack-name> <keywords>*
//   <pack-name> ::= <word> | "-"
//   <keywords> ::= <word>*
// If pack-name != "-", SearchGifs will limit results to gifs from pack-name, otherwise SearchGifs will search in all
// packs that user is subscribed to.
//
// If there is no pack called pack-name, SearchGifs will return no results.
// If <keywords> is provided, SearchGifs will filter the gifs it returns to only those containing <keywords>.
func SearchGifs(ctx context.Context, user int, query string) ([]Gif, error) {
	// no results for an empty query
	if query == "" {
		return nil, nil
	}

	split := collapseWhitespaceRegex.Split(query, -1)
	packName := split[0]
	var keywords []string
	if len(split) > 1 {
		keywords = split[1:]
	}

	var results []Gif
	if packName == "-" {
		// get all packs user is subscribed to
		q := datastore.NewQuery(subscriptionKind).Filter("User =", user)

		var subscriptions []Subscription
		_, err := q.GetAll(ctx, &subscriptions)
		if err != nil {
			return nil, err
		}

		// if the user is not subscribed to any packs, return an empty slice and no error
		if len(subscriptions) == 0 {
			return nil, nil
		}

		// search for gifs
		gIndex, err := search.Open(gifsIndex)
		if err != nil {
			return nil, err
		}

		for _, s := range subscriptions {
			var q string
			if len(keywords) > 0 {
				q = fmt.Sprintf("Pack = %s AND Keywords = (%s)", s.Pack, strings.Join(keywords, " OR "))
			} else {
				q = fmt.Sprintf("Pack = %s", s.Pack)
			}

			for t := gIndex.Search(ctx, q, nil); ; {
				var gif Gif
				_, err := t.Next(&gif)
				if err != nil {
					if err == search.Done {
						break
					} else {
						return nil, err
					}
				}

				results = append(results, gif)
			}
		}
	} else {
		// return nil if packName is invalid
		if !packNameRegex.MatchString(packName) {
			return nil, nil
		}

		var pack Pack
		key := datastore.NewKey(ctx, packKind, packName, 0, nil)
		err := datastore.Get(ctx, key, &pack)
		if err != nil {
			if err == datastore.ErrNoSuchEntity {
				return nil, nil
			} else {
				return nil, err
			}
		}

		// search for gifs
		gIndex, err := search.Open(gifsIndex)
		if err != nil {
			return nil, err
		}

		var q string
		if len(keywords) > 0 {
			q = fmt.Sprintf("Pack = %s AND Keywords = (%s)", packName, strings.Join(keywords, " OR "))
		} else {
			q = fmt.Sprintf("Pack = %s", packName)
		}

		for t := gIndex.Search(ctx, q, nil); ; {
			var gif Gif
			_, err := t.Next(&gif)
			if err != nil {
				if err == search.Done {
					break
				} else {
					return nil, err
				}
			}

			results = append(results, gif)
		}
	}

	return results, nil
}
