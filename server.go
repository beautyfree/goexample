package main

import (
  "os"
  "fmt"
  "time"
  "log"
  "net/http"
  "path"
  "strconv"

  "github.com/go-martini/martini"
  "github.com/igm/pubsub"
  "github.com/carbocation/go-instagram/instagram"
  "gopkg.in/igm/sockjs-go.v2/sockjs"
  gooauth2 "github.com/golang/oauth2"
  "github.com/martini-contrib/oauth2"
  "github.com/martini-contrib/sessions"
)

var chat pubsub.Publisher

// Returns a new Instagram OAuth 2.0 backend endpoint.
func Instagram(opts *gooauth2.Options) martini.Handler {
  authUrl := "https://api.instagram.com/oauth/authorize"
  tokenUrl := "https://api.instagram.com/oauth/access_token"
  return oauth2.NewOAuth2Provider(opts, authUrl, tokenUrl)
}

func main() {
  go check()

  handler := sockjs.NewHandler("/echo", sockjs.DefaultOptions, echoHandler)

  m := martini.Classic();
  m.NotFound(func(w http.ResponseWriter, r *http.Request) {
    // Only rewrite paths *not* containing filenames
    if path.Ext(r.URL.Path) == "" {
      http.ServeFile(w, r, "public/index.html")
    } else {
      w.WriteHeader(http.StatusNotFound)
      w.Write([]byte("404 page not found"))
    }
  })
  m.Use(sessions.Sessions("my_session", sessions.NewCookieStore([]byte("secret123"))))
  m.Use(Instagram(&gooauth2.Options{
    ClientID: "04df7554ee08464da70ee2530cc84774",
    ClientSecret: "cb8ae1c1ec8f4fd98dba157ebf3b0d8b",
    RedirectURL: "http://go.ngrok.com/oauth2callback",
  }))
  m.Use(func(c martini.Context, tokens oauth2.Tokens, s sessions.Session) {
    if !tokens.IsExpired() {
      log.Println(tokens.Access())
      //log.Println(s.Get(tokens.Access()))
      log.Println(tokens.ExtraData())
    }
    c.Next()
  })

  m.Use(martini.Logger())
  m.Use(martini.Recovery())
  m.Use(martini.Static("public"))

/*
  m.Get("/", oauth2.LoginRequired, func(tokens oauth2.Tokens) string {
    http.ServeFile(w, r, "public/index.html")
  })*/

  /*
  m.Get("/", func() string {
    return "Hello world!"
  })
  */

  http.Handle("/echo/", handler)
  http.Handle("/", m)
  http.ListenAndServe(":3000", nil)
}


func echoHandler(session sockjs.Session) {
  log.Println("new sockjs session established")
  var closedSession = make(chan struct{})
  chat.Publish("[info] new participant joined chat")
  defer chat.Publish("[info] participant left chat")
  go func() {
  	reader, _ := chat.SubChannel(nil)
  	for {
  		select {
  		case <-closedSession:
  			return
  		case msg := <-reader:
  			if err := session.Send(msg.(string)); err != nil {
  				return
  			}
  		}

  	}
  }()
  for {
  	if msg, err := session.Recv(); err == nil {
  		chat.Publish(msg)
  		continue
  	}
  	break
  }
  close(closedSession)
  log.Println("sockjs session closed")
}


func check() {
  // You can optionally pass your own HTTP's client, otherwise pass it with nil.
  client := instagram.NewClient(nil)
  client.ClientID = "8f2c0ad697ea4094beb2b1753b7cde9c"

  //wait around with a forever loop
  for {
    time.Sleep(1 * time.Second)

    media, err := client.Media.Get("822463308874820376")
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    }
    log.Println(media.Likes.Count)
    chat.Publish(strconv.Itoa(media.Likes.Count))
  }
}
