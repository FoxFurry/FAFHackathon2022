package server

import (
	"context"
	"fmt"
	"net/http"

	"fafhackathon2022/internal/http/jwt"
	"fafhackathon2022/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const (
	authSchema = "Bearer "
)

type Server interface {
	Start()
	Stop()
}

type server struct {
	service service.Service
	jwt     jwt.JWTProvider
}

func New(ctx context.Context) Server {
	//audience := []string{viper.GetString("AUTH0_AUDIENCE")}
	return &server{
		service: service.New(),
		jwt:     jwt.NewJWT(viper.GetString("AUTH0_DOMAIN"), []byte("FE55n9uZsJZ9MeEE1Vk965r2HhmnT0du")),
	}
}

func (s *server) Start() {
	router := gin.Default()

	router.GET("/login", s.corsMiddleware, s.jwtMiddleware, s.Login)
	router.GET("/ws/:token", s.WS)

	router.GET("/login/test", s.corsMiddleware, s.randJWTMiddleware, s.Login)
	router.GET("/ws/test/:uuid", s.WSTest)

	router.Run(fmt.Sprintf(":%s", viper.GetString("PORT")))
}

func (s *server) Stop() {

}

func (s *server) randJWTMiddleware(c *gin.Context) {
	c.Set("uuid", c.GetHeader("test_uuid"))
}

func (s *server) jwtMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	if len(authHeader) <= len(authSchema) {
		c.String(http.StatusUnauthorized, "missing or invalid JWT token")
		c.Abort()
		return
	}

	token := authHeader[len(authSchema):]

	uuid, err := s.jwt.ValidateToken(token)
	if err != nil {
		c.String(http.StatusUnauthorized, fmt.Sprintf("could not validate JWT token: %v", err))
		c.Abort()
		return
	}

	c.Set("uuid", uuid)
	c.Next()
}

func (s *server) corsMiddleware(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204)
		return
	}

	c.Next()
}
