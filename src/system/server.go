package system

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/nicolasvancan/monvandb/src/database"
)

const MessageDelimiter = "\nfimdaparada\n"

// Connection stores client connection details
type Connection struct {
	Conn     net.Conn
	User     User
	Database *database.Database
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
	creds := string(buf[:n])
	username, password, dbName, err := parseCredentials(creds)

	if err != nil {
		conn.Write([]byte("Invalid credentials format"))
		return
	}

	// Get database
	db, err := database.GetDatabase(dbName)

	if err != nil {
		conn.Write([]byte("Database not found"))
		return
	}

	// Get user by name
	user, err := GetUser(username)
	if err != nil {
		conn.Write([]byte("User not found"))
		return
	}

	err = user.ValidatePassword(password)
	if err != nil {
		conn.Write([]byte("Invalid password"))
		return
	}

	// Store connection
	s.mu.Lock()
	s.connections[username] = &Connection{
		Conn:     conn,
		User:     *user,
		Database: db,
	}
	s.mu.Unlock()

	// Send welcome message (encrypted by TLS)
	_, err = conn.Write([]byte("Welcome to the RDBMS!"))

	if err != nil {
		fmt.Println("Error sending welcome message:", err)
		return
	}
	var message []byte
	// Handle further communication
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Connection closed:", err)
			break
		}
		fmt.Println("Received message:", string(buf[:n]))
		// Prcess further messages as needed
		message := append(message, buf[:n]...)

		// Read last bytes and compare whether or not they are equal to the
		// Message delimiter
		if string(buf[len(MessageDelimiter)-1:]) == MessageDelimiter {
			// Process the message
			fmt.Println("Processing message:", string(message))
			message = make([]byte, 0)
		}
	}
}

func parseCredentials(creds string) (string, string, string, error) {
	// Placeholder: Implement parsing of "username:password:database"
	// In production, use a structured format (e.g., JSON)
	// mvdb://username:password@database

	initialSubstr := "mvdb://"
	if len(creds) < len(initialSubstr) || creds[:len(initialSubstr)] != initialSubstr {
		return "", "", "", errors.New("invalid credentials format")
	}

	creds = creds[len(initialSubstr):]
	parts := strings.Split(creds, "@")
	if len(parts) != 2 {
		return "", "", "", errors.New("invalid credentials format")
	}

	credentials := strings.Split(parts[0], ":")
	if len(credentials) != 2 {
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
