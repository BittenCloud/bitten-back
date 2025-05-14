package interfaces

type SQLDatabase interface {
	Connect()
	Shutdown()
}
