package netlab

import "fmt"

func GetVethNames(name string) []string {
	return []string{
		fmt.Sprintf("%s-0", name),
		fmt.Sprintf("%s-1", name),
	}
}
