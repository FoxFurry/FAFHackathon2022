package server

import (
	"fmt"
	"log"
	"net/http"

	"fafhackathon2022/internal/http/connections"
	"fafhackathon2022/internal/store/models"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *server) Login(c *gin.Context) {
	userUUID := c.GetString("uuid")
	if userUUID == "" {
		c.JSON(http.StatusInternalServerError, "unable to extract user uuid")
		return
	}

	userData, err := s.service.GetUser(c, userUUID)
	if err != nil {
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
	}

	c.JSON(200, userData)
}

func (s *server) WS(c *gin.Context) {
	token := c.Param("token")

	userUUID, err := s.jwt.ValidateToken(token)
	if err != nil {
		c.String(http.StatusUnauthorized, err.Error())
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("unable to upgrade to websocket: %v", err))
		return
	}

	connections.SetConnection(userUUID, ws)

	defer ws.Close()
	for {
		var message models.Message
		err := ws.ReadJSON(&message)
		if err != nil {
			log.Printf("websocket connection interrupted: %v", err)
			return
		}

		resp, err := s.service.HandleWSMessage(c, userUUID, message)
		if err != nil {
			ws.WriteJSON(gin.H{
				"message": err.Error(),
			})
		} else {
			ws.WriteJSON(resp)
		}
	}
}

func (s *server) WSTest(c *gin.Context) {
	userUUID := c.Param("uuid")

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("unable to upgrade to websocket: %v", err))
		return
	}

	connections.SetConnection(userUUID, ws)

	defer func(ws *websocket.Conn, userUUID string) {
		ws.Close()
		connections.DeleteConnection(userUUID)

	}(ws, userUUID)
	for {
		var message models.Message
		err := ws.ReadJSON(&message)
		if err != nil {
			log.Printf("websocket connection interrupted: %v", err)
			return
		}

		resp, err := s.service.HandleWSMessage(c, userUUID, message)
		if err != nil {
			ws.WriteJSON(gin.H{
				"message": err.Error(),
			})
		} else {
			ws.WriteJSON(resp)
		}
	}
}
