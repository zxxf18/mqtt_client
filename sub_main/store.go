package main

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/zxxf18/mqtt_client/utils"
)

const (
	table = `
CREATE TABLE IF NOT EXISTS gitbug_message (
	id bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'primary key',
	topic varchar(128) NOT NULL DEFAULT '' COMMENT 'topic',
	content text NOT NULL COMMENT 'content',
	create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'create time',
	update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'update time',
	PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='Gitbug message table';
`
)

type DBConfig struct {
	Database struct {
		Type            string `yaml:"type" json:"type" validate:"nonzero"`
		URL             string `yaml:"url" json:"url" validate:"nonzero"`
		MaxConns        int    `yaml:"maxConns" json:"maxConns" default:20`
		MaxIdleConns    int    `yaml:"maxIdleConns" json:"maxIdleConns" default:5`
		ConnMaxLifetime int    `yaml:"connMaxLifetime" json:"connMaxLifetime" default:150`
	} `yaml:"database" json:"database" default:"{}"`
}

type DB struct {
	db  *sqlx.DB
	cfg DBConfig
}

func NewDB(path string) (*DB, error) {
	var cfg DBConfig
	err := utils.LoadYAML(path, &cfg)
	if err != nil {
		return nil, err
	}
	db, err := sqlx.Open(cfg.Database.Type, cfg.Database.URL)
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetMaxOpenConns(cfg.Database.MaxConns)
	db.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &DB{
		db:  db,
		cfg: cfg,
	}, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) Save(topic string, content []byte) error {
	sql := fmt.Sprintf(`INSERT INTO gitbug_message (topic, content) VALUES (?, ?)`)
	_, err := d.db.Exec(sql, topic, base64.StdEncoding.EncodeToString(content))
	if err != nil {
		return err
	}
	return nil
}
