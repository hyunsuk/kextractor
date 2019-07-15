package ask

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func confirm(question, ok, cancel string) (bool, error) {
	var response string
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(question)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	response = strings.Trim(response, " \n")
	if response != ok && response != cancel {
		return confirm(question, ok, cancel)
	}
	if response == cancel {
		return false, nil
	}
	return true, nil
}

// Confirm .
func Confirm(question, ok, cancel string) (bool, error) {
	return confirm(question, ok, cancel)
}
