package driver

var Drivers = map[string]SiteDriver{}

type SiteDriver interface {
	PutFile(localPrefix string, path string, force bool) (err error) // Upload the file at localPrefix/path to /path on the site
	FlushFiles(validPaths []string) (err error)                      // Flush any files from site not included in the slice of validPaths
}
