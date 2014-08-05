package testutils

import (
	"archive/tar"
	"fmt"

	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type TarFileContent struct {
	Name, Body string
}

func GenerateTarfile(fs boshsys.FileSystem, tarFileContents []TarFileContent) (string, error) {
	tempFile, err := fs.TempFile("bmtestutil")
	if err != nil {
		return "", err
	}

	//DEBUG
	fmt.Println("tempFile: ", tempFile.Name())

	tarWriter := tar.NewWriter(tempFile)

	for _, tarFileContent := range tarFileContents {
		hdr := &tar.Header{
			Name: tarFileContent.Name,
			Size: int64(len(tarFileContent.Body)),
			Mode: 0644,
		}

		err = tarWriter.WriteHeader(hdr)
		if err != nil {
			return "", err
		}

		_, err = tarWriter.Write([]byte(tarFileContent.Body))
		if err != nil {
			return "", err
		}
	}

	err = tarWriter.Close()
	if err != nil {
		return "", err
	}

	err = tempFile.Close()
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}
