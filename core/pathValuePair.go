package core

// PathValuePair encapsulates a config path (eg. data.storageClass) and the value that it has.
// Used during the 'set' command to store parsed config paths and values.
type PathValuePair struct {
	Path  []string
	Value string
}
