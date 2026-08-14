//go:build windows

package keyring

import (
	"syscall"
	"unsafe"
)

var (
	modAdvapi32     = syscall.NewLazyDLL("advapi32.dll")
	procCredWriteW  = modAdvapi32.NewProc("CredWriteW")
	procCredReadW   = modAdvapi32.NewProc("CredReadW")
	procCredDeleteW = modAdvapi32.NewProc("CredDeleteW")
	procCredFree    = modAdvapi32.NewProc("CredFree")
)

const (
	credTypeGeneric       = 1
	credPersistMachine    = 2
	credErrNotFound       = syscall.Errno(1168)
	errKeyringUnsupported = "OS keyring backend is not supported on this platform"
)

type credentialW struct {
	Flags              uint32
	Type               uint32
	TargetName         *uint16
	Comment            *uint16
	LastWritten        syscall.Filetime
	CredentialBlobSize uint32
	CredentialBlob     *byte
	Persist            uint32
	AttributeCount     uint32
	Attributes         uintptr
	TargetAlias        *uint16
	UserName           *uint16
}

func supportsKeyring() bool {
	return true
}

func credWrite(target string, blob []byte) error {
	targetPtr, err := syscall.UTF16PtrFromString(target)
	if err != nil {
		return err
	}
	cred := credentialW{
		Type:       credTypeGeneric,
		TargetName: targetPtr,
		Persist:    credPersistMachine,
	}
	if len(blob) > 0 {
		cred.CredentialBlobSize = uint32(len(blob))
		cred.CredentialBlob = &blob[0]
	}
	r1, _, callErr := procCredWriteW.Call(uintptr(unsafe.Pointer(&cred)), 0)
	if r1 == 0 {
		return normalizeCredErr(callErr)
	}
	return nil
}

func credRead(target string) ([]byte, error) {
	targetPtr, err := syscall.UTF16PtrFromString(target)
	if err != nil {
		return nil, err
	}
	var pcred *credentialW
	r1, _, callErr := procCredReadW.Call(
		uintptr(unsafe.Pointer(targetPtr)),
		uintptr(credTypeGeneric),
		0,
		uintptr(unsafe.Pointer(&pcred)),
	)
	if r1 == 0 {
		return nil, normalizeCredErr(callErr)
	}
	defer procCredFree.Call(uintptr(unsafe.Pointer(pcred)))
	if pcred == nil || pcred.CredentialBlobSize == 0 || pcred.CredentialBlob == nil {
		return nil, nil
	}
	blob := make([]byte, int(pcred.CredentialBlobSize))
	copy(blob, unsafe.Slice(pcred.CredentialBlob, int(pcred.CredentialBlobSize)))
	return blob, nil
}

func credDelete(target string) error {
	targetPtr, err := syscall.UTF16PtrFromString(target)
	if err != nil {
		return err
	}
	r1, _, callErr := procCredDeleteW.Call(uintptr(unsafe.Pointer(targetPtr)), uintptr(credTypeGeneric), 0)
	if r1 == 0 {
		if callErr == credErrNotFound {
			return nil
		}
		return normalizeCredErr(callErr)
	}
	return nil
}

func normalizeCredErr(err error) error {
	if errno, ok := err.(syscall.Errno); ok {
		if errno == 0 {
			return syscall.EINVAL
		}
		return errno
	}
	if err == nil {
		return syscall.EINVAL
	}
	return err
}

func credNotFound(err error) bool {
	errno, ok := err.(syscall.Errno)
	return ok && errno == credErrNotFound
}
