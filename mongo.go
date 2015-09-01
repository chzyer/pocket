package main

import (
	"labix.org/v2/mgo"
)

var (
	globalSession *mgo.Session
	globalDbname  string
)

type Session struct {
	s *mgo.Session
	*mgo.Database
}

func (s *Session) Close() {
	s.s.Close()
}

func InitMongo(url_, db string) {
	session, err := mgo.Dial(url_)
	if err != nil {
		panic(err)
	}
	globalDbname = db
	globalSession = session
}

func Mongo() *Session {
	s := globalSession.Copy()
	return &Session{
		s:        s,
		Database: s.DB(globalDbname),
	}
}
