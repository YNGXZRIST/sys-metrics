package config

func GetConfigArgs(args []string) []string {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-c", "-config", "--config":
			if i+1 < len(args) {
				return args[i : i+2]
			}
			return args[i : i+1]
		}
	}

	return nil
}
