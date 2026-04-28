package errorx

type Error interface {
	error
	Code() int
	I18nKey() string
}
