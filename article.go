package main

import (
	"fmt"
	"net/url"
	"time"

	"gopkg.in/logex.v1"
	"gopkg.in/mgo.v2/bson"
)

const ArticleName = "Article"

type Article struct {
	Id       bson.ObjectId `bson:"_id"`
	Title    string
	Host     string
	Url      string
	Deleted  bool
	ReadTime time.Time
	Source   []byte
	Gen      []byte
}

func (a *Article) Link() string {
	u, _ := url.Parse(a.Url)
	u.Scheme = ""
	return u.String()[2:]
}

func NewArticle(url_, title string, source, gen []byte) *Article {
	u, _ := url.Parse(url_)
	return &Article{
		Id:     bson.NewObjectId(),
		Title:  title,
		Host:   u.Host,
		Url:    url_,
		Source: source,
		Gen:    gen,
	}
}

func FindArticles(s *Session) (a []*Article) {
	err := s.C(ArticleName).Find(bson.M{
		"deleted": bson.M{"$ne": true},
	}).Sort("-readtime", "-_id").All(&a)
	if err != nil {
		logex.Error(err)
	}
	return
}

func FindArticle(s *Session, url_ string) (a *Article) {
	err := s.C(ArticleName).Find(bson.M{
		"url":     url_,
		"deleted": bson.M{"$ne": true},
	}).One(&a)
	if err != nil {
		logex.Error(err)
	}
	return
}

func DeleteArticle(s *Session, id string) error {
	if !bson.IsObjectIdHex(id) {
		return fmt.Errorf("invalid id")
	}
	return s.C(ArticleName).UpdateId(bson.ObjectIdHex(id), bson.M{
		"$set": bson.M{"deleted": true},
	})
}

func (a *Article) Save(s *Session) error {
	_, err := s.C(ArticleName).UpsertId(a.Id, bson.M{"$set": a})
	return err
}
