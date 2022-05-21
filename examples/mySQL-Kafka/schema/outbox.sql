CREATE TABLE outbox (
        id varchar(100) NOT NULL,
        data BLOB NOT NULL,
        state INT NOT NULL,
        created_on DATETIME NOT NULL,
        locked_by varchar(100) NULL,
        locked_on DATETIME NULL,
        processed_on DATETIME NULL,
        number_of_attempts INT NOT NULL,
        last_attempted_on DATETIME NULL,
        error varchar(1000) NULL
)