package ycq

//import (
//"net/http"
//"net/http/httptest"

//"github.com/jetbasrawi/goes"
//. "gopkg.in/check.v1"
//)

//var _ = Suite(&GEventStoreSuite{})

//var (
//mux *http.ServeMux

//server *httptest.Server

//client *goes.Client
//)

//func setupSimulator(es []*goes.Event, m *goes.Event) {
//u, _ := url.Parse(server.URL)
//handler := goes.ESAtomFeedSimulator{Events: es, BaseURL: u, MetaData: m}
//mux.Handle("/", handler)
//}

//type GEventStoreSuite struct {
//eventBus   *InternalEventBus
//eventStore *GetEventStore
//client     *goes.Client
//}

//func (s *GEventStoreSuite) SetUpTest(c *C) {
//mux = http.NewServeMux()
//server = httptest.NewServer(mux)
//s.client, _ = goes.NewClient(nil, server.URL)
//s.eventBus = NewInternalEventBus()
//s.eventStore = NewGetEventStore(s.eventBus, client, client, client)
//}

//func (s *GEventStoreSuite) TestNewEventStore(c *C) {
//es := NewGetEventStore(s.eventBus, client, client, client)
//c.Assert(es.eventBus, Equals, s.eventBus)
//}

//func (s *GEventStoreSuite) TestCanSetEventFactory(c *C) {
//eventFactory := NewDelegateEventFactory()
//s.eventStore.SetEventFactory(eventFactory)
//c.Assert(s.eventStore.eventFactory, Equals, eventFactory)
//}

//type SimEventType struct {
//Foo string `json:"foo"`
//}

//func (s *GEventStoreSuite) TestReadFromSim(c *C) {
//stream := "some-stream"
//es := goes.CreateTestEvents(10, stream, server.URL, "SimEventType")
//setupSimulator(es, nil)

//eventBus := NewInternalEventBus()

//eventStore := NewGetEventStore(eventBus, s.client, s.client, s.client)
//eventFactory := NewDelegateEventFactory()
//eventFactory.RegisterDelegate(&SimEventType{}, func() interface{} { return &SimEventType{} })
//eventStore.SetEventFactory(eventFactory)

//got, err := eventStore.Load(stream)
//c.Assert(err, IsNil)

//spew.Dump(got)
//}
