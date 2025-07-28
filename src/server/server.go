package server

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/nicolasvancan/monvandb/src/database"
	"github.com/nicolasvancan/monvandb/src/dbvengine"
	"github.com/nicolasvancan/monvandb/src/dbvengine/dataframe"
	"github.com/nicolasvancan/monvandb/src/system"
)

const MessageDelimiter = "d2a81ac4c97aa94e0c3474c619cc85508a19544f388c9c1c0d26773a2b19eb09"

var globalContext *dbvengine.Context

func init() {
	// Initialize global context
	// Load system users and roles into memory
	globalContext = &dbvengine.Context{
		Ttl:              600 * 3,
		ExecutionResults: make(map[string]dataframe.Dataframe),
	}
}

// Connection stores client connection details
type Connection struct {
	Conn     net.Conn
	User     system.User
	Database *database.Database
	Engine   *dbvengine.VirtualEngine
}

// Server holds the connection map
type Server struct {
	connections map[string]*Connection
	Port        int
	mu          sync.Mutex
}

func NewServer(port int) *Server {
	return &Server{
		connections: make(map[string]*Connection),
		Port:        port,
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read username and password (encrypted by TLS)
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err)
		return
	}
	creds := buf[:n]
	username, password, dbName, err := parseCredentials(creds)

	if err != nil {
		conn.Write(BuildFailureMessage("Invalid credentials format"))
		return
	}

	// Get database
	db, err := database.GetDatabase(dbName)

	if err != nil {
		conn.Write(BuildFailureMessage("Database not found"))
		return
	}

	// Get user by name
	user, err := system.GetUser(username)
	if err != nil {
		conn.Write(BuildFailureMessage("User not found"))
		return
	}

	err = user.ValidatePassword(password)
	if err != nil {
		conn.Write(BuildFailureMessage("Invalid password"))
		return
	}

	// Store connection
	s.mu.Lock()
	s.connections[username] = &Connection{
		Conn:     conn,
		User:     *user,
		Database: db,
		Engine:   dbvengine.NewVirtualEngine(globalContext),
	}
	s.mu.Unlock()

	// Send welcome message (encrypted by TLS)
	_, err = conn.Write(BuildSuccessMessage("Welcome to the RDBMS!", 0))

	if err != nil {
		fmt.Println("Error sending welcome message:", err)
		return
	}

	var message []byte = make([]byte, 0)
	// Handle further communication
	for {
		buf := make([]byte, 1024)
		print("Waiting for message from client...\n")
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Connection closed:", err)
			break
		}
		// Prcess further messages as needed
		message = append(message, buf[:n]...)

		// Read last bytes and compare whether or not they are equal to the
		// Message delimiter

		if len(message) > len(MessageDelimiter)+1 {
			if string(message[len(message)-len(MessageDelimiter):]) == MessageDelimiter {
				// Process the message
				parsedMessage, err := extractJSONField(message)
				if err != nil {
					conn.Write(BuildFailureMessage(err.Error()))
					message = make([]byte, 0)
					continue
				}

				if parsedMessage.RequestType == CLOSE_CONN {
					s.mu.Lock()
					delete(s.connections, username)
					s.mu.Unlock()
					conn.Close()
					return
				}

				if parsedMessage.RequestType != QUERY {
					conn.Write(BuildFailureMessage("Invalid request type " + string(parsedMessage.RequestType)))
					message = make([]byte, 0)
					continue
				}

				// Process query
				queryMessage, ok := parsedMessage.Message.(string)
				if !ok {
					conn.Write(BuildFailureMessage("Invalid query message type, required string"))
					message = make([]byte, 0)
					continue
				}

				if queryMessage == "" {
					conn.Write(BuildFailureMessage("Empty query message"))
					message = make([]byte, 0)
					continue
				}

				executionResults := s.connections[username].Engine.Execute(s.connections[username].Database.Name, queryMessage)
				if executionResults.Status != dbvengine.ExecutionSuccess {
					conn.Write(BuildFailureMessage("Error executing query: " + executionResults.ErrorMessage))
					message = make([]byte, 0)
					continue
				}

				conn.Write(BuildSuccessMessage(executionResults.Result.ToJsonObject(), executionResults.Duration))
				message = make([]byte, 0)
			}
		}
	}
}

func extractJSONField(incommingMessage []byte) (UserMessage, error) {
	// Placeholder: Implement JSON extraction logic
	// In production, use a proper JSON library
	if len(incommingMessage) == 0 {
		return UserMessage{}, errors.New("empty message")
	}
	fmt.Println("Received message:", string(incommingMessage))
	fmt.Println(string(incommingMessage[:len(incommingMessage)-len(MONVAN_EOF)]))
	var parsedJson UserMessage
	err := json.Unmarshal(incommingMessage[:len(incommingMessage)-len(MONVAN_EOF)], &parsedJson)
	// Example parsing logic (to be replaced with actual JSON parsing)
	if err != nil {
		return UserMessage{}, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return parsedJson, nil
}

func parseCredentials(creds []byte) (string, string, string, error) {
	// Placeholder: Implement parsing of "username:password:database"
	// In production, use a structured format (e.g., JSON)
	// mvdb://username:password@database

	parsedJson, err := extractJSONField(creds)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse credentials: %w", err)
	}

	if parsedJson.RequestType != AUTHENTICATE {
		return "", "", "", errors.New("invalid request type - expected 'authenticate'")
	}

	credMessage := parsedJson.Message.(string)

	initialSubstr := "mvdb://"
	if len(credMessage) < len(initialSubstr) || credMessage[:len(initialSubstr)] != initialSubstr {
		fmt.Println("Invalid credentials format first", credMessage)
		return "", "", "", errors.New("invalid credentials format")
	}

	credMessage = credMessage[len(initialSubstr):]
	parts := strings.Split(credMessage, "@")
	if len(parts) != 2 {
		fmt.Println("Invalid credentials format split @", credMessage)
		return "", "", "", errors.New("invalid credentials format")
	}

	credentials := strings.Split(parts[0], ":")
	if len(credentials) != 2 {
		fmt.Println("Invalid credentials format split :", credMessage)
		return "", "", "", errors.New("invalid credentials format")
	}

	username := credentials[0]
	password := credentials[1]
	database := parts[1]
	if database == "" {
		return "", "", "", errors.New("database name is required")
	}

	return username, password, database, nil
}

func (s *Server) Run() {
	// Load server certificate and key
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		fmt.Println("Error loading certificate:", err)
		return
	}

	// Configure TLS
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12, // Enforce modern TLS versions
	}

	// Start TLS listener
	listener, err := tls.Listen("tcp", ":"+strconv.Itoa(s.Port), config)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("TLS Server listening on :8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go s.handleConnection(conn)
	}
}
