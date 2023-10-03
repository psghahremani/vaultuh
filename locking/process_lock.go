package locking

import (
	"errors"
	"fmt"
	"github.com/psghahremani/vaultuh/utility"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
)

var lockFilePath = fmt.Sprintf("%s/vaultuh.lock", os.TempDir())

type ProcessLock struct{}

func (f ProcessLock) Hold() (bool, error) {
	file, err := os.OpenFile(
		lockFilePath,
		os.O_RDWR|os.O_CREATE|os.O_EXCL,
		0b110000000,
	)
	if err == nil {
		_, err = file.WriteString(fmt.Sprintf("%d\n", os.Getpid()))
		if err != nil {
			_ = file.Close()
			return false, fmt.Errorf("could not write into the lock file: %w", err)
		}
		err = file.Sync()
		if err != nil {
			_ = file.Close()
			return false, fmt.Errorf("could not \"fsync\" the lock file: %w", err)
		}
		err = file.Close()
		if err != nil {
			return false, fmt.Errorf("could not close the lock file: %w", err)
		}
		return true, nil
	}
	if err.(*os.PathError).Err == syscall.EEXIST {
		isStray, err := f.checkIfHeldLockIsStray()
		if err != nil {
			return false, fmt.Errorf("could not check if the currently held lock is stray: %w", err)
		}
		if !isStray {
			return false, nil
		}
		_, err = utility.DeleteFile(lockFilePath)
		if err != nil {
			return false, fmt.Errorf("could not release the stray lock: %w", err)
		}
		return f.Hold()
	}
	return false, fmt.Errorf("could not create the lock file: %w", err)
}

func (f ProcessLock) checkIfHeldLockIsStray() (bool, error) {
	file, err := os.OpenFile(
		lockFilePath,
		os.O_RDONLY,
		0b100000000,
	)
	if err != nil {
		return false, fmt.Errorf("could not open the lock file: %w", err)
	}

	contents, err := io.ReadAll(file)
	if err != nil {
		return false, fmt.Errorf("could not read the lock file: %w", err)
	}
	parts := strings.Split((string)(contents), "\n")
	if len(parts) != 2 {
		return false, fmt.Errorf("could not parse the lock file: %w", errors.New("expected 1 \"NewLine\" character at the end of the file"))
	}

	processID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return false, fmt.Errorf("detected an invalid PID in the lock file: %w", err)
	}
	process, err := os.FindProcess((int)(processID))
	if err != nil {
		return false, fmt.Errorf("could not find a process by its PID: %w", err)
	}

	err = process.Signal(syscall.Signal(0))
	if err == nil {
		return false, nil
	}
	if errors.Is(err, os.ErrProcessDone) {
		return true, nil
	}
	errorCode, isValid := err.(syscall.Errno)
	if isValid {
		if errorCode == syscall.ESRCH {
			return true, nil
		}
		if errorCode == syscall.EPERM {
			return false, nil
		}
	}
	return false, fmt.Errorf("could not send a signal to the process: %w", err)
}

func (f ProcessLock) Release() error {
	_, err := utility.DeleteFile(lockFilePath)
	if err != nil {
		return fmt.Errorf("could not delete the lock file: %w", err)
	}
	return nil
}
