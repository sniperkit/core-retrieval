package core

import (
	"database/sql"
	"io/ioutil"
	"os"

	"gopkg.in/src-d/core-retrieval.v0/model"
	"gopkg.in/src-d/core-retrieval.v0/repository"

	"gopkg.in/src-d/framework.v0/configurable"
	"gopkg.in/src-d/framework.v0/database"
	"gopkg.in/src-d/framework.v0/queue"
	"gopkg.in/src-d/go-billy.v3"
	"gopkg.in/src-d/go-billy.v3/osfs"
)

const transactionerLocalDir = "transactioner"

type containerConfig struct {
	configurable.BasicConfiguration
	TempDir             string `default:"/tmp/sourced"`
	Broker              string `default:"amqp://localhost:5672"`
	RootRepositoriesDir string `default:"/tmp/root-repositories"`
}

var config = &containerConfig{}

func init() {
	configurable.InitConfig(config)
}

var container struct {
	Broker               queue.Broker
	Database             *sql.DB
	ModelRepositoryStore *model.RepositoryStore
	ModelMentionStore    *model.MentionStore
	RootedTransactioner  repository.RootedTransactioner
	TempDirFilesystem    billy.Filesystem
}

// Broker returns a queue.Broker for the default queue system.
func Broker() queue.Broker {
	if container.Broker == nil {
		b, err := queue.NewBroker(config.Broker)
		if err != nil {
			panic(err)
		}

		container.Broker = b
	}

	return container.Broker
}

// Database returns a sql.DB for the default database. If it is not possible to
// connect to the database, this function will panic. Multiple calls will always
// return the same instance.
func Database() *sql.DB {
	if container.Database == nil {
		container.Database = database.Must(database.Default())
	}

	return container.Database
}

// ModelMentionStore returns the default *model.ModelMentionStore, using the
// default database. If it is not possible to connect to the database, this
// function will panic. Multiple calls will always return the same instance.
func ModelMentionStore() *model.MentionStore {
	if container.ModelMentionStore == nil {
		container.ModelMentionStore = model.NewMentionStore(Database())
	}

	return container.ModelMentionStore
}

// ModelRepositoryStore returns the default *model.RepositoryStore, using the
// default database. If it is not possible to connect to the database, this
// function will panic. Multiple calls will always return the same instance.
func ModelRepositoryStore() *model.RepositoryStore {
	if container.ModelRepositoryStore == nil {
		container.ModelRepositoryStore = model.NewRepositoryStore(Database())
	}

	return container.ModelRepositoryStore
}

// TemporaryFilesystem returns a billy.Filesystem that can be used to store
// temporary files. This directory is dedicated to the running application.
func TemporaryFilesystem() billy.Filesystem {
	if container.TempDirFilesystem == nil {
		if err := os.MkdirAll(config.TempDir, os.FileMode(0755)); err != nil {
			panic(err)
		}

		dir, err := ioutil.TempDir(config.TempDir, "")
		if err != nil {
			panic(err)
		}

		container.TempDirFilesystem = osfs.New(dir)
	}

	return container.TempDirFilesystem
}

// RootedTransactioner returns the default RootedTransactioner instance,
// using the default RootRepositories directory. The local filesystem used to
// create the transactioner is the default TemporaryFilesystem from core container.
func RootedTransactioner() repository.RootedTransactioner {
	if container.RootedTransactioner == nil {
		tmpFs, err := TemporaryFilesystem().Chroot(transactionerLocalDir)
		if err != nil {
			panic(err)
		}

		container.RootedTransactioner =
			repository.NewSivaRootedTransactioner(
				osfs.New(config.RootRepositoriesDir),
				tmpFs,
			)
	}

	return container.RootedTransactioner
}
