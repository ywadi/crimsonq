package Defs

const (
	STATUS_PENDING   string = "pending"
	STATUS_ACTIVE    string = "active"
	STATUS_DELAYED   string = "delayed"
	STATUS_FAILED    string = "failed"
	STATUS_COMPLETED string = "completed"
)

const (
	CREATED_AT   string = "created_at"
	PENDING_AT   string = "pending_at"
	DELAYED_AT   string = "delayed_at"
	FAILED_AT    string = "failed_at"
	COMPLETED_AT string = "completed_at"
)

const (
	QDB_PREFIX string = "qdbid:"
)

const (
	ERRincorrectConsumerId    string = "001:incorrect_consumer_id"
	ERRnoDataReturn           string = "002:no_data_returned"
	ERRIncorrectStatus        string = "003:incorrect_status"
	ERROConsumerAlreadyExists string = "004:consumer_exists"
	ERRExceededConcurrency    string = "005:pull_exceeded_concurrency"
	ERRIncorrectMessageId     string = "006:incorrect message id"
	ERRStatusNotAllowed       string = "007:method_not_allowed_on_http"
	ERRIncorrectArgs          string = "008: incorrect_args"
)

const (
	HTTP_GET  string = "GET"
	HTTP_POST string = "POST "
)
