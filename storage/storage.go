package storage

type Storage interface {
	// Bootstrap creates the table or record in the storage
	// source for later consumption or an error if it wasn't possible.
	Bootstrap(key string) error

	// Get gets the value for the given key from the storage source
	Get(key string) (string, error)

	// Set sets the value for the given key in the storage source
	Set(key, value string) error

	// IsValidKey checks if the key used for storage is valid
	// for the storage engine source
	IsValidKey(key string) error

	// ConfigString returns a stringified version of the current
	// configuration for logging purposes
	ConfigString() string
}
