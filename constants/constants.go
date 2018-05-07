package constants

//TrueString string version of true boolean
const TrueString = "true"

//SeedFileName defines the filename for the seed file
const SeedFileName = "seed.manifest.json"

//DefaultRegistry defines the default registry address to use when searching for images
const DefaultRegistry = "https://hub.docker.com/"

//DefaultOrg defines the default organization to use when searching for images
const DefaultOrg = "geoint"

//SchemaType defines manfiest or metadata
type SchemaType int

const (
	//SchemaManifest manifest schema
	SchemaManifest SchemaType = iota

	//SchemaMetadata metadata schema
	SchemaMetadata
)

//DockerConfigDir defines directory to use for DOCKER_CONFIG environment variable
//This is used instead of the default directory so when seed is run as root (most times),
//user credentials aren't stored under the root directory and people aren't stepping on
//each other
const DockerConfigDir = "docker-config-"

const DockerConfigKey = "DOCKER_CONFIG"