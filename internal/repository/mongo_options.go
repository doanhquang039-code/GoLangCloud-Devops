package repository

import "go.mongodb.org/mongo-driver/mongo/options"

func optionsReplaceUpsert() *options.ReplaceOptions {
	return options.Replace().SetUpsert(true)
}
