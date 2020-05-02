package pool

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	pwdFilePath    = "/etc/node-pwd/node-passwd"
	pwdTmpFilePath = "/etc/node-pwd/node-passwd.tmp"
)

func replacePassword(nodePasswordMap map[string]string) error {
	shouldReplace, err := func(nodepwd map[string]string) (bool, error) {
		file, err := os.Open(pwdFilePath)
		if err != nil {
			return false, err
		}
		defer file.Close()

		tmpFile, err := os.Create(pwdTmpFilePath)
		if err != nil {
			return false, err
		}
		defer tmpFile.Close()

		reader := bufio.NewReader(file)
		writter := bufio.NewWriter(tmpFile)
		shouldReplace := false

		for {
			line, _, err := reader.ReadLine()
			if err == io.EOF {
				break
			}
			parts := strings.Split(string(line), ",")
			if len(parts) != 4 {
				continue
			}
			pwd := parts[0]
			name := parts[1]

			gotPwd, ok := nodepwd[name]
			if ok && gotPwd != pwd {
				logrus.Infof("Password mismatch line {%s}", string(line))
				shouldReplace = true
			} else {
				writter.WriteString(string(line) + "\n")
			}
		}

		return shouldReplace, writter.Flush()
	}(nodePasswordMap)

	if err != nil {
		return errors.Errorf("Error creating tmp password file: %s", err)
	}

	if shouldReplace {
		err := func() error {
			file, err := os.OpenFile(pwdFilePath, os.O_WRONLY, os.ModeAppend)
			if err != nil {
				return errors.Errorf("Error replacing password file: %s", err)
			}

			defer file.Close()

			tmpFile, err := os.Open(pwdTmpFilePath)
			if err != nil {
				return errors.Errorf("Error replacing password file: %s", err)
			}

			defer tmpFile.Close()
			_, err = io.Copy(file, tmpFile)
			if err != nil {
				return err
			}
			// TODO remove this restart
			logrus.Infof("Successfully replaced password file. Restarting ...")
			os.Exit(0)
			return nil
		}()

		if err != nil {
			return errors.Errorf("Error replacing password file: %s", err)
		}
		return nil
	}
	return nil
}
