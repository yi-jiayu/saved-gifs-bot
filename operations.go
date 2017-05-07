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
	listKind         = "List"
	subscriptionKind = "Subscription"
)

var (
	collapseWhitespaceRegex = regexp.MustCompile(`\s+`)
	listNameRegex           = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
)

// exported errors
var (
	ErrInvalidName = errors.New("invalid name")
	ErrNotAllowed  = errors.New("not allowed")
	ErrNotFound    = errors.New("list not found")
)

type Gif struct {
	List     search.Atom
	FileID   search.Atom
	Keywords string
}

type List struct {
	Name    string
	Creator string
}

type Subscription struct {
	User string
	List string
}

// Returns true if list was created, false if a list with the same name already exists.
func NewList(ctx context.Context, name, creator string) (bool, error) {
	// validate list name
	if !listNameRegex.MatchString(name) {
		return false, ErrInvalidName
	}

	// check if name name is already taken
	var list List
	key := datastore.NewKey(ctx, listKind, name, 0, nil)
	err := datastore.Get(ctx, key, &list)
	if err != nil {
		if err != datastore.ErrNoSuchEntity {
			return false, err
		}
	} else {
		return false, nil
	}

	list = List{
		Name:    name,
		Creator: creator,
	}

	_, err = datastore.Put(ctx, key, &list)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Returns a slice of lists which were created by creator
func MyLists(ctx context.Context, creator string) ([]List, error) {
	q := datastore.NewQuery(listKind).Filter("Creator =", creator)

	var lists []List
	_, err := q.GetAll(ctx, &lists)
	if err != nil {
		return nil, err
	}

	return lists, nil
}

// Returns true if user was successfully subscribed to list, false if user was already subscribed to list.
// Returns ErrNotFound if list does not exist.
func Subscribe(ctx context.Context, list, user string) (bool, error) {
	// validate list name
	if !listNameRegex.MatchString(list) {
		return false, ErrInvalidName
	}

	// check if list exists
	var l List
	key := datastore.NewKey(ctx, listKind, list, 0, nil)
	err := datastore.Get(ctx, key, &l)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return false, ErrNotFound
		} else {
			return false, err
		}
	}

	// check if user is already subscribed
	q := datastore.NewQuery(subscriptionKind).Filter("User =", user).Filter("List =", list)

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
		List: list,
	}

	key = datastore.NewIncompleteKey(ctx, subscriptionKind, nil)
	_, err = datastore.Put(ctx, key, &subscription)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Returns true if user was successfully unsubscribed from list, false if user was not subscribed to list.
// err will be ErrInvalidName if list is not a valid list name
func Unsubscribe(ctx context.Context, list, user string) (bool, error) {
	// validate list name
	if !listNameRegex.MatchString(list) {
		return false, ErrInvalidName
	}

	// check if user is already subscribed
	q := datastore.NewQuery(subscriptionKind).Filter("User =", user).Filter("List =", list)

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

func MySubscriptions(ctx context.Context, user string) ([]Subscription, error) {
	q := datastore.NewQuery(subscriptionKind).Filter("User =", user)

	var subscriptions []Subscription
	_, err := q.GetAll(ctx, &subscriptions)
	if err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func NewGif(ctx context.Context, list, user string, gif Gif) error {
	// validate list name
	if !listNameRegex.MatchString(list) {
		return ErrInvalidName
	}

	// check that user is the creator of list
	var l List
	key := datastore.NewKey(ctx, listKind, list, 0, nil)
	err := datastore.Get(ctx, key, &l)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return ErrNotFound
		} else {
			return err
		}
	}

	if user != l.Creator {
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

func SearchGifs(ctx context.Context, user, query string) ([]Gif, error) {
	// get all lists user is subscribed to
	q := datastore.NewQuery(subscriptionKind).Filter("User =", user)

	var subscriptions []Subscription
	_, err := q.GetAll(ctx, &subscriptions)
	if err != nil {
		return nil, err
	}

	// if the user is not subscribed to any lists, return an empty slice and no error
	if len(subscriptions) == 0 {
		return nil, nil
	}

	// search for gifs
	gIndex, err := search.Open(gifsIndex)
	if err != nil {
		return nil, err
	}

	var results []Gif
	for _, s := range subscriptions {
		q := fmt.Sprintf("List = %s AND Keywords = (%s)", s.List, strings.Join(collapseWhitespaceRegex.Split(query, -1), " OR "))
		for u := gIndex.Search(ctx, q, nil); ; {
			var g Gif
			_, err := u.Next(&g)
			if err != nil {
				if err == search.Done {
					break
				} else {
					return nil, err
				}
			}

			results = append(results, g)
		}
	}

	return results, nil
}
