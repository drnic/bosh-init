package tar

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type cmdExtractor struct {
	runner boshsys.CmdRunner
	fs     boshsys.FileSystem
	logger boshlog.Logger
}

func NewCmdExtractor(runner boshsys.CmdRunner, fs boshsys.FileSystem, logger boshlog.Logger) cmdExtractor {
	return cmdExtractor{
		runner: runner,
		fs:     fs,
		logger: logger,
	}
}

func (e cmdExtractor) Extract(path string) (string, error) {
	extractedPath, err := e.fs.TempDir("tar-cmdExtractor")
	if err != nil {
		return "", bosherr.WrapError(err, "Creating tempdir for tar extraction")
	}

	_, _, _, err = e.runner.RunCommand("tar", "-C", extractedPath, "-xzf", path)
	if err != nil {
		e.fs.RemoveAll(extractedPath)
		return "", bosherr.WrapError(err, "Extracting tar '%s'", path)
	}

	return extractedPath, nil
}
