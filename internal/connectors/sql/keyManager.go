package sql

import "bitback/internal/interfaces"

type KeyManagerRepo struct {
	db interfaces.SQLDatabase
}
