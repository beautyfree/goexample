package main

import (
  "log"

  "github.com/go-martini/martini"
  
  gooauth2 "github.com/golang/oauth2"
  "github.com/martini-contrib/oauth2"
  "github.com/martini-contrib/sessions"
)

// Returns a new Instagram OAuth 2.0 backend endpoint.
func Instagram(opts *gooauth2.Options) martini.Handler {
  authUrl := "https://api.instagram.com/oauth/authorize"
  tokenUrl := "https://api.instagram.com/oauth/access_token"
  return oauth2.NewOAuth2Provider(opts, authUrl, tokenUrl)
}

func main() {
  m := martini.Classic();
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
  m.Get("/", func() string {
    return "Hello world!"
  })
  m.Run()
}

