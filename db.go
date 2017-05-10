package main

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

// Gif represents a gif in our search index
type Gif struct {
	Pack     search.Atom
	FileID   search.Atom
	Keywords string
}

// Pack represents a gif pack in datastore
type Pack struct {
	Name         string
	Creator      int
	Contributors []int
}

// Subscription represents a subscription to a gif pack in datastore
type Subscription struct {
	UserID int
	Pack   string
}

// UserPacks represents the packs a user has created and is a contributor to
type UserPacks struct {
	IsCreator     []Pack
	IsContributor []Pack
}

// NewPack returns true if pack was created, false if a pack with the same name already exists.
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

// GetUserPacks returns a UserPacks struct representing the packs a user has created and is a contributor to
func GetUserPacks(ctx context.Context, userID int) (UserPacks, error) {
	var isCreator []Pack
	q1 := datastore.NewQuery(packKind).Filter("Creator =", userID)
	_, err := q1.GetAll(ctx, &isCreator)
	if err != nil {
		return UserPacks{}, err
	}

	var isContributor []Pack
	q2 := datastore.NewQuery(packKind).Filter("Contributors =", userID)
	_, err = q2.GetAll(ctx, &isContributor)
	if err != nil {
		return UserPacks{}, err
	}

	userPacks := UserPacks{
		IsCreator:     isCreator,
		IsContributor: isContributor,
	}

	return userPacks, nil
}

// GetPack retrieves information about a specific gif pack by the pack name
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
		}

		return Pack{}, err
	}

	return pack, nil
}

// SetPack is a convenience wrapper around datastore.Put to update the value of pack in the datastore
func SetPack(ctx context.Context, pack *Pack) error {
	// validate pack name
	if !packNameRegex.MatchString(pack.Name) {
		return ErrInvalidName
	}

	key := datastore.NewKey(ctx, packKind, pack.Name, 0, nil)
	_, err := datastore.Put(ctx, key, pack)
	if err != nil {
		return err
	}

	return nil
}

// NewContributor adds a contributor to a gif pack
func NewContributor(ctx context.Context, packName string, creator, contributor int) (bool, error) {
	// validate pack name
	if !packNameRegex.MatchString(packName) {
		return false, ErrInvalidName
	}

	// check that pack exists and creator is the creator of pack
	pack, err := GetPack(ctx, packName)
	if err != nil {
		return false, err
	}

	if creator != pack.Creator {
		return false, ErrNotAllowed
	}

	// check that contributor is not already in pack
	for _, c := range pack.Contributors {
		if contributor == c {
			return false, nil
		}
	}

	// update pack
	pack.Contributors = append(pack.Contributors, contributor)
	err = SetPack(ctx, &pack)
	if err != nil {
		return false, err
	}

	return true, nil
}

// DeleteContributor removes a contributor from a gif pack
func DeleteContributor(ctx context.Context, packName string, creator, contributor int) (bool, error) {
	// validate pack name
	if !packNameRegex.MatchString(packName) {
		return false, ErrInvalidName
	}

	// check that creator is the creator of pack
	pack, err := GetPack(ctx, packName)
	if err != nil {
		return false, err
	}

	if creator != pack.Creator {
		return false, ErrNotAllowed
	}

	// check that contributor is in pack
	var index int
	found := false
	for i, c := range pack.Contributors {
		if contributor == c {
			index = i
			found = true
			break
		}
	}

	if !found {
		return false, nil
	}

	// remove contributor from pack.Contributors
	a := pack.Contributors
	a[index] = a[len(a)-1]

	// update pack
	pack.Contributors = a[:len(a)-1]
	err = SetPack(ctx, &pack)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Subscribe returns true if user was successfully subscribed to pack, false if user was already subscribed to pack.
// err will be ErrNotFound if pack does not exist.
func Subscribe(ctx context.Context, pack string, userID int) (bool, error) {
	// validate pack name
	if !packNameRegex.MatchString(pack) {
		return false, ErrInvalidName
	}

	// check if pack exists
	_, err := GetPack(ctx, pack)
	if err != nil {
		return false, err
	}

	// check if userID is already subscribed
	q := datastore.NewQuery(subscriptionKind).Filter("UserID =", userID).Filter("Pack =", pack).Limit(1).KeysOnly()

	keys, err := q.GetAll(ctx, nil)
	if err != nil {
		return false, err
	}

	if len(keys) > 0 {
		return false, nil
	}

	subscription := Subscription{
		UserID: userID,
		Pack:   pack,
	}

	key := datastore.NewIncompleteKey(ctx, subscriptionKind, nil)
	_, err = datastore.Put(ctx, key, &subscription)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Unsubscribe returns true if user was successfully unsubscribed from pack, false if user was not subscribed to pack.
// err will be ErrInvalidName if pack is not a valid pack name
func Unsubscribe(ctx context.Context, pack string, user int) (bool, error) {
	// validate pack name
	if !packNameRegex.MatchString(pack) {
		return false, ErrInvalidName
	}

	// check if user is already subscribed
	q := datastore.NewQuery(subscriptionKind).Filter("UserID =", user).Filter("Pack =", pack).Limit(1).KeysOnly()

	var keys []*datastore.Key
	keys, err := q.GetAll(ctx, nil)
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

// MySubscriptions returns a slice of the subscriptions a user has.
func MySubscriptions(ctx context.Context, user int) ([]Subscription, error) {
	q := datastore.NewQuery(subscriptionKind).Filter("UserID =", user)

	var subscriptions []Subscription
	_, err := q.GetAll(ctx, &subscriptions)
	if err != nil {
		return nil, err
	}

	return subscriptions, nil
}

// GetGif is a convenience wrapper to get a gif by packName and fileID from the search index
func GetGif(ctx context.Context, packName, fileID string) (Gif, error) {
	index, err := search.Open(gifsIndex)
	if err != nil {
		return Gif{}, err
	}

	var gif Gif
	key := fmt.Sprintf("%s:%s", packName, fileID)
	err = index.Get(ctx, key, &gif)
	if err != nil {
		if err == search.ErrNoSuchDocument {
			return Gif{}, ErrNotFound
		}

		return Gif{}, err
	}

	return gif, nil
}

// HasEditPermissions returns true if userID is the creator of pack or a contributor to pack.
func HasEditPermissions(pack Pack, userID int) bool {
	if userID == pack.Creator {
		return true
	}

	for _, c := range pack.Contributors {
		if userID == c {
			return true
		}
	}

	return false
}

// NewGif adds a new gif to pack. Returns true if a new gif was added to the pack, false if that gif was already in the
// pack.
func NewGif(ctx context.Context, packName string, userID int, gif Gif) (bool, error) {
	// validate pack name
	if !packNameRegex.MatchString(packName) {
		return false, ErrInvalidName
	}

	// check if user is the creator or a contributor to pack
	pack, err := GetPack(ctx, packName)
	if err != nil {
		return false, err
	}

	if !HasEditPermissions(pack, userID) {
		return false, ErrNotAllowed
	}

	// check if gif is already in pack
	_, err = GetGif(ctx, packName, string(gif.FileID))
	if err != nil {
		if err != ErrNotFound {
			return false, err
		}
	} else {
		return false, nil
	}

	// add gif to pack
	index, err := search.Open(gifsIndex)
	if err != nil {
		return false, err
	}

	key := fmt.Sprintf("%s:%s", packName, gif.FileID)
	_, err = index.Put(ctx, key, &gif)
	if err != nil {
		return false, err
	}

	return true, nil
}

// EditGif updates a gif's keywords. Returns true if the gif existed and was updated, false if the gif did not exist in
// pack.
func EditGif(ctx context.Context, packName string, userID int, gif Gif) (bool, error) {
	// validate pack name
	if !packNameRegex.MatchString(packName) {
		return false, ErrInvalidName
	}

	// check if user is the creator or a contributor to pack
	pack, err := GetPack(ctx, packName)
	if err != nil {
		return false, err
	}

	if !HasEditPermissions(pack, userID) {
		return false, ErrNotAllowed
	}

	// check if gif is already in pack
	_, err = GetGif(ctx, packName, string(gif.FileID))
	if err != nil {
		if err == ErrNotFound {
			return false, nil
		}

		return false, err
	}

	// update gif
	index, err := search.Open(gifsIndex)
	if err != nil {
		return false, err
	}

	key := fmt.Sprintf("%s:%s", packName, gif.FileID)
	_, err = index.Put(ctx, key, &gif)
	if err != nil {
		return false, err
	}

	return true, nil
}

// DeleteGif removes a gif from pack. Returns true if the the gif was deleted from the pack, false if the gif was not
// part of the pack.
func DeleteGif(ctx context.Context, packName string, userID int, fileID string) (bool, error) {
	// validate pack name
	if !packNameRegex.MatchString(packName) {
		return false, ErrInvalidName
	}

	// check that user is the creator of pack
	pack, err := GetPack(ctx, packName)
	if err != nil {
		return false, err
	}

	if !HasEditPermissions(pack, userID) {
		return false, ErrNotAllowed
	}

	// check if gif is already in pack
	index, err := search.Open(gifsIndex)
	if err != nil {
		return false, err
	}

	var g Gif
	key := fmt.Sprintf("%s:%s", packName, fileID)
	err = index.Get(ctx, key, &g)
	if err != nil {
		if err == search.ErrNoSuchDocument {
			return false, nil
		}

		return false, err
	}

	err = index.Delete(ctx, key)
	if err != nil {
		return false, err
	}

	return true, nil
}

// SearchGifs returns gifs from the search index matching query.
//
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
	// default to searching all subscribed packs
	if query == "" {
		query = "-"
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
		q := datastore.NewQuery(subscriptionKind).Filter("UserID =", user)

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
			}

			return nil, err
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
