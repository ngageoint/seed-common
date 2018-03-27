package util

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ngageoint/seed-common/constants"
)

//MachineCheckSudo Checks error for telltale sign seed command should be run as sudo
func MachineCheckSudo(machine string) {
	cmd := exec.Command("docker-machine", "ssh", machine, "docker", "info")

	// attach stderr pipe
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		PrintUtil("ERROR: Error attaching to version command stderr. %s\n", err.Error())
	}

	// Run docker build
	if err := cmd.Start(); err != nil {
		PrintUtil("ERROR: Error executing docker version. %s\n",
			err.Error())
	}

	slurperr, _ := ioutil.ReadAll(errPipe)
	er := string(slurperr)
	if er != "" {
		if strings.Contains(er, "Cannot connect to the Docker daemon. Is the docker daemon running on this host?") ||
			strings.Contains(er, "dial unix /var/run/docker.sock: connect: permission denied") {
			PrintUtil("Elevated permissions are required by seed to run Docker. Try running the seed command again as sudo.\n")
			panic(Exit{1})
		}
	}
}

//MachineDockerVersionHasLabel returns if the docker version is greater than 1.11.1
func MachineDockerVersionHasLabel(machine string) bool {
	return MachineDockerVersionGreaterThan(machine, 1, 11, 1)
}

//MachineDockerVersionHasReferenceFilter returns if the docker version is greater than 1.13.0
func MachineDockerVersionHasReferenceFilter(machine string) bool {
	return MachineDockerVersionGreaterThan(machine, 1, 13, 0)
}

//MachineDockerVersionGreaterThan returns if the docker version is greater than the specified version
func MachineDockerVersionGreaterThan(machine string, major, minor, patch int) bool {
	cmd := exec.Command("docker-machine", "ssh", machine, "docker", "version", "-f", "{{.Client.Version}}")

	// Attach stdout pipe
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		PrintUtil("ERROR: Error attaching to version command stdout. %s\n", err.Error())
	}

	// Run docker version
	if err := cmd.Start(); err != nil {
		PrintUtil("ERROR: Error executing docker version. %s\n", err.Error())
	}

	// Print out any std out
	slurp, _ := ioutil.ReadAll(outPipe)
	if string(slurp) != "" {
		version := strings.Split(string(slurp), ".")

		// check each part of version. Return false if 1st < 1, 2nd < 11, 3rd < 1
		if len(version) > 1 {
			v1, _ := strconv.Atoi(version[0])
			v2, _ := strconv.Atoi(version[1])

			// check for minimum of 1.11.1
			if v1 == major {
				if v2 > minor {
					return true
				} else if v2 == minor && len(version) == 3 {
					v3, _ := strconv.Atoi(version[2])
					if v3 >= patch {
						return true
					}
				}
			} else if v1 > major {
				return true
			}

			return false
		}
	}

	return false
}

//MachineImageExists returns true if a local image already exists, false otherwise
func MachineImageExists(machine, imageName string) (bool, error) {
	// Test if image has been built; Rebuild if not
	imgsArgs := []string{"ssh", machine, "docker", "images", "-q", imageName}
	imgOut, err := exec.Command("docker-machine", imgsArgs...).Output()
	if err != nil {
		PrintUtil("ERROR: Error executing docker %v\n", imgsArgs)
		PrintUtil("%s\n", err.Error())
		return false, err
	} else if string(imgOut) == "" {
		PrintUtil("INFO: No docker image found locally for image name %s.\n",
			imageName)
		return false, nil
	}
	return true, nil
}

//MachineSaveImage Saves specified image on given machine
func MachineSaveImage(machine, imageName string) (string, error) {
	// Remove the tag name
	imageTar := strings.Split(imageName, ":")[0] + ".tar"
	args := []string{"ssh", machine, "docker", "save", "-o", imageTar, imageName}
	out, err := exec.Command("docker-machine", args...).Output()
	if err != nil {
		PrintUtil("ERROR: Error executing docker-machine %v\n", args)
		PrintUtil("%s\n", err.Error())
		return "", err
	} else if string(out) != "" {
		PrintUtil("ERROR: Error saving image %s.\n%s\n",
			imageName, string(out))
		return "", errors.New("ERROR saving image to " + imageTar + ".\n")
	}
	return imageTar, nil
}

//MachineSCP SCPs the specified image file to the given machine
func MachineSCP(file, machinePath, machine string) (bool, error) {
	node := machine + ":"+machinePath
	args := []string{"scp", file, node}
	out, err := exec.Command("docker-machine", args...).Output()
	if err != nil {
		PrintUtil("ERROR: Error executing docker-machine %v\n", args)
		PrintUtil("%s\n", err.Error())
		return false, err
	} else if string(out) != "" {
		PrintUtil("ERROR: Error SCP of %s.\n%s\n",
			file, string(out))
		return false, errors.New("ERROR transferring image to " + machine + " via SCP.\n")
	}
	return true, nil
}

//MachineLoad loads the specified image file onto the given machine
func MachineLoad(file, machine string) (bool, error) {
	imgFile := "/tmp/"+file
	args := []string{"ssh", machine, "docker", "load", "-i", imgFile}
	out, err := exec.Command("docker-machine", args...).Output()
	if err != nil {
		PrintUtil("ERROR: Error executing docker-machine %v\n", args)
		PrintUtil("%s\n", err.Error())
		return false, err
	} else if string(out) == "" {
		PrintUtil("ERROR: Error loading image from %s.\n",
			imgFile)
		return false, errors.New("ERROR: Error executing docker-machine %v\n"+
			"Error loading image to " + machine + ".\n")
	}
	return true, nil
}

//MachineLogin logs into the specified registry on the given machine
func MachineLogin(machine, registry, username, password string) error {
	var errs, out bytes.Buffer
	args := []string{"ssh", machine, "docker", "login", "-u", username, "-p", password, registry}
	cmd := exec.Command("docker-machine", args...)
	cmd.Stderr = io.MultiWriter(os.Stderr, &errs)
	cmd.Stdout = &out

	err := cmd.Run()


	errStr := strings.ToUpper(errs.String())
	if strings.Contains(errStr, "WARNING") {
		//report warnings but don't return error (i.e. --password via CLI is insecure warning)
		PrintUtil("Docker login warning: %s\n", errs.String())
	}

	if strings.Contains(errStr, "ERROR") {
		PrintUtil("ERROR: Error reading stderr %s\n", errs.String())
		return errors.New(errs.String())
	}

	if err != nil {
		errMsg := fmt.Sprintf("ERROR: Error executing docker login.\n%s\n", err.Error())
		err = errors.New(errMsg)
		return err
	}

	PrintUtil("%s", out.String())
	return nil
}

//MachineTag tags an image
func MachineTag(machine, origImg, img string) error {
	var errs bytes.Buffer

	// Run docker tag
	if img != origImg {
		tagCmd := exec.Command("docker-machine", "ssh", machine, "docker", "tag", origImg, img)
		tagCmd.Stderr = io.MultiWriter(os.Stderr, &errs)
		tagCmd.Stdout = os.Stderr

		if err := tagCmd.Run(); err != nil {
			PrintUtil("ERROR: Error executing docker tag. %s\n",
				err.Error())
		}
		if errs.String() != "" {
			PrintUtil("ERROR: Error tagging image '%s':\n%s\n", origImg, errs.String())
			PrintUtil("Exiting seed...\n")
			return errors.New(errs.String())
		}
	}

	return nil
}

//MachinePush pushes an image to its repository
func MachinePush(machine, img string) error {
	var errs bytes.Buffer

	// docker push
	args := []string{"ssh", machine, "docker", "push", img}
	pushCmd := exec.Command("docker-machine", args...)
	pushCmd.Stderr = io.MultiWriter(os.Stderr, &errs)
	pushCmd.Stdout = os.Stdout

	// Run docker push
	if err := pushCmd.Run(); err != nil {
		PrintUtil("ERROR: Error executing docker push. %s\n",
			err.Error())
		return err
	}

	// Check for errors. Exit if error occurs
	if errs.String() != "" {
		PrintUtil("ERROR: Error pushing image '%s':\n%s\n", img,
			errs.String())
		PrintUtil("Exiting seed...\n")
		return errors.New(errs.String())
	}

	return nil
}

//MachineRemoveImage removes the specified image from the machine
func MachineRemoveImage(machine, img string) error {
	var errs bytes.Buffer

	PrintUtil("INFO: Removing local image %s\n", img)
	rmiCmd := exec.Command("docker-machine", "ssh", machine, "docker", "rmi", img)
	rmiCmd.Stderr = io.MultiWriter(os.Stderr, &errs)
	rmiCmd.Stdout = os.Stdout

	if err := rmiCmd.Run(); err != nil {
		PrintUtil("ERROR: Error executing docker rmi. %s\n",
			err.Error())
		return err
	}

	// check for errors on stderr
	if errs.String() != "" {
		PrintUtil("ERROR: Error removing image '%s':\n%s\n", img,
			errs.String())
		PrintUtil("Exiting seed...\n")
		return errors.New(errs.String())
	}

	return nil
}

//MachineCreateRegistry creates a registry with the given port on the specified machine
func MachineCreateRegistry(machine string, port int) (bool, error) {

	// Set default port if not provided
	if port < 0 {
		port = 5000
	}
	portStr := strconv.Itoa(port) + ":" + strconv.Itoa(port)
	// args := []string{"ssh", clusterMaster, "docker", "service", "create", "--name", "registry", "--publish", portStr, "registry:2"}
	args := []string{"ssh", machine, "docker", "run", "-d", "-p", portStr, "--name", "registry", "registry:2"}
	out, err := exec.Command("docker-machine", args...).Output()
	if err != nil {
		PrintUtil("ERROR: Error creating registry service.\n%s\n", err.Error())
		return false, err

	} else if strings.Contains(string(out), "rpc error") {
		// Already a registry service available.. proceeding
		if strings.Contains(string(out), "name conflicts with an existing object") {
			return true, nil
		} 
		PrintUtil("ERROR: Port %s is already in use by another service. Please use another port number.\n", string(port))
		return false, errors.New("Port " + string(port) + " is already in use by another service. Please use another port number.")
	}

	return true, nil
}

//MachineRestartRegistry restarts the registry
func MachineRestartRegistry() error {
	PrintUtil("RESTARTING REGISTRY........................\n.\n.\n.\n.\n.\n")
	var errs bytes.Buffer

	PrintUtil("INFO: Restarting test registry...\n")
	cmd := exec.Command("../restartRegistry.sh")
	cmd.Stderr = io.MultiWriter(os.Stderr, &errs)
	cmd.Stdout = os.Stdout

	err := cmd.Run()

	// check for errors on stderr first; it will likely have more explanation than cmd.Run
	if errs.String() != "" {
		PrintUtil("ERROR: Error restarting registry. %s\n", errs.String())
		PrintUtil("Exiting seed...\n")
		return errors.New(errs.String())
	}

	if err != nil {
		PrintUtil("ERROR: Error restarting registry. %s\n",
			err.Error())
		return err
	}

	return nil
}

//MachinePull pulls specified image from remote repository (default docker.io)
//returns the name of the remote image retrieved, if any
func MachinePull(machine, image, registry, org, username, password string) (string, error) {
	if username != "" {
		//set config dir so we don't stomp on other users' logins with sudo
		configDir := constants.DockerConfigDir + time.Now().Format(time.RFC3339)
		os.Setenv(constants.DockerConfigKey, configDir)
		defer RemoveAllFiles(configDir)
		defer os.Unsetenv(constants.DockerConfigKey)

		err := Login(registry, username, password)
		if err != nil {
			fmt.Println(err)
			return "", err
		}
	}

	if registry == "" {
		registry = constants.DefaultRegistry
	}

	registry = strings.Replace(registry, "https://hub.docker.com", "docker.io", 1)

	remoteImage := fmt.Sprintf("%s/%s", registry, image)

	if org != "" {
		remoteImage = fmt.Sprintf("%s/%s/%s", registry, org, image)
	}

	var errs, out bytes.Buffer
	// pull image
	pullArgs := []string{"ssh", machine, "docker", "pull", remoteImage}
	pullCmd := exec.Command("docker-machine", pullArgs...)
	pullCmd.Stderr = io.MultiWriter(os.Stderr, &errs)
	pullCmd.Stdout = &out

	err := pullCmd.Run()
	if err != nil {
		PrintUtil("ERROR: Error executing docker pull.\n%s\n", err.Error())
		return "", err
	}

	if errs.String() != "" {
		PrintUtil("ERROR: Error reading stderr %s\n", errs.String())
		return "", errors.New(errs.String())
	}

	return remoteImage, nil
}

//MachineGetSeedManifestFromImage returns the manifest of the given image
func MachineGetSeedManifestFromImage(machine, imageName string) (string, error) {
	cmdStr := "inspect -f '{{index .Config.Labels \"com.ngageoint.seed.manifest\"}}'" + imageName
	PrintUtil("INFO: Retrieving seed manifest from %s LABEL=com.ngageoint.seed.manifest\n", imageName)

	inspectCommand := exec.Command("docker-machine", "ssh", machine, "docker", "inspect", "-f",
		"'{{index .Config.Labels \"com.ngageoint.seed.manifest\"}}'", imageName)

	errPipe, err := inspectCommand.StderrPipe()
	if err != nil {
		PrintUtil("ERROR: error attaching to docker inspect command stderr. %s\n", err.Error())
	}

	// Attach stdout pipe
	outPipe, err := inspectCommand.StdoutPipe()
	if err != nil {
		PrintUtil("ERROR: error attaching to docker inspect command stdout. %s\n", err.Error())
	}

	// Run docker inspect
	if err = inspectCommand.Start(); err != nil {
		PrintUtil("ERROR: error executing docker %s. %s\n", cmdStr, err.Error())
	}

	// Print out any std out
	seedBytes, err := ioutil.ReadAll(outPipe)
	if err != nil {
		fmt.Fprintf(os.Stdout, "ERROR: Error retrieving docker %s stdout.\n%s\n",
			cmdStr, err.Error())
	}

	// check for errors on stderr
	slurperr, _ := ioutil.ReadAll(errPipe)
	if string(slurperr) != "" {
		PrintUtil("ERROR: Error executing docker %s:\n%s\n",
			cmdStr, string(slurperr))
	}

	// un-escape special characters
	label := string(seedBytes)

	seedStr := UnescapeManifestLabel(label)

	return seedStr, err
}

//MachineHome returns the machine's home directory
func MachineHome(machine string) string {
	homeBytes, err := exec.Command("docker-machine", "ssh", machine, "printf", "$(printenv | sed -n -e 's/HOME=//p')").Output()
	if err != nil {
		PrintUtil("ERROR: Error retrieving machine's home directory\n%s", err.Error())
		return ""	}
	home := filepath.Join(string(filepath.Separator), string(homeBytes))

	return home
}

//Mount mounts the specified path to the $HOME directory of the specified machine
// Creates the path on the specified machine if it does not exist
// Must mount to the home directory to avoid permission issues
// Assumes the path e
func Mount(machine, machinePath, path string) error {
	// Create path if doesn't exist
	_, err := exec.Command("docker-machine", "ssh", machine, "test", "-d", machinePath).Output()
	if err != nil {
		_, err = exec.Command("docker-machine", "ssh", machine, "mkdir", "-p", machinePath).Output()
		if err != nil {
			PrintUtil("ERROR: Error creating %s on %s\n", machinePath, machine)
			return err
		}
	}

	// Perform mount
	args := []string{"mount", machine+":"+machinePath, path}
	out, err := exec.Command("docker-machine", args...).Output()
	if err != nil  {
		PrintUtil("ERROR: Error mounting %s\n%s\n", path, err.Error())
		return err
	} else if string(out) != "" {
		PrintUtil("ERROR: Error mountingg %s\n%s\n", path, string(out))
		return errors.New(string(out))
	}
	return nil
}

//Unmount unmounts path from machine
func Unmount(machine, machinepath, path string) error {

	// Try 'docker-machine mount -u'
	if _, err := exec.Command("docker-machine", "mount", "-u", machine+":"+machinepath, path).Output(); err != nil {
		err = nil
		if _, err := exec.Command("umount", path).Output(); err != nil {
			PrintUtil("Error unmounting %s\n%s\n", path, err.Error())
			return err
		}
	}
	return  nil
}