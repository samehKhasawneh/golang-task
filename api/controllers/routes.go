package controllers

import (
	"golang-task/api/middlewares"
)

func (s *Server) initializeRoutes() {

	api := s.Router.Group("/api")
	{
		// Login Route
		api.POST("/login", s.Login)
		api.POST("/logout", s.Logout)

		//Users routes
		api.POST("/register", s.CreateUser)

		//Tweets routes
		api.GET("/tweets", middlewares.TokenAuthMiddleware(), s.SearchTweets)
		api.POST("/tweets", middlewares.TokenAuthMiddleware(), s.SaveTweets)
		api.POST("/posts", middlewares.TokenAuthMiddleware(), s.CreateTweet)
		api.GET("/posts", middlewares.TokenAuthMiddleware(), s.GetTweets)
		api.GET("/posts/:id", middlewares.TokenAuthMiddleware(), s.GetTweet)
		api.PUT("/posts/:id", middlewares.TokenAuthMiddleware(), s.UpdateTweet)
		api.DELETE("/posts/:id", middlewares.TokenAuthMiddleware(), s.DeleteTweet)
		api.GET("/user_posts/:id", middlewares.TokenAuthMiddleware(), s.GetUserTweets)
	}
}
