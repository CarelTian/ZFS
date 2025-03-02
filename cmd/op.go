package cmd

type Manager struct {
	nodeName     string
	relativePath string
}

func interpret(command string, zfs *Manager) {
	//parts := strings.Split(command, " ")

}

func (zfs *Manager) GetFile(filename string) bool {

	return true
}
