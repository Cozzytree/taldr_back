package server

import (
	"fmt"
	"net/http"

	"github.com/Cozzytree/taldrBack/internal/models"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	WriteBufferSize: 1024,
	ReadBufferSize:  1024,
	CheckOrigin: func(r *http.Request) bool {
		fmt.Println("socket host :", r.Host)
		// Allow connections from any origin (you may want to restrict this in production)
		return true
	},
}

func (s *Server) WsConnect(w http.ResponseWriter, r *http.Request) {
	// Upgrade the connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Only send the error if it hasn't been written already
		http.Error(w, "error while upgrading to websocket", http.StatusUpgradeRequired)
		return
	}
	defer conn.Close()

	type input struct {
		Type     string `json:"type"`
		Id       string `json:"id"`
		UserId   string `json:"userId"`
		Document string `json:"document"`
		models.ShapeProps
	}
	var workspaceDetails input

	for {

		// Read data from the WebSocket connection
		err := conn.ReadJSON(&workspaceDetails)
		if err != nil {
			break
		}

		switch workspaceDetails.Type {
		case "doc":
			s.shapeS.
				saveDocument(fmt.Sprintf("%v+%v", workspaceDetails.Id, workspaceDetails.UserId),
					workspaceDetails.Document)
			break
		case "canvas":
			s.shapeS.
				storeShapes(fmt.Sprintf("%v+%v", workspaceDetails.Id, workspaceDetails.UserId),
					workspaceDetails.Shapes)
			conn.WriteJSON(map[string]string{"status": "success"})
			break
		default:
			conn.WriteJSON(map[string]string{"status": "unknown"})
		}
	}
}
