package statusCode

const (
	Failed                         = 0
	Success                        = 1
	DBError                        = 102
	NoApplicationNotAllowOperation = 103
	MissingMandatoryItem           = 104
	EmailCodeError                 = 105
	BuildJwtError                  = 106
	SendEmailError                 = 107
	HasBindTotp                    = 108
	ValidateTOTPError              = 109
	ParseMultipartFormError        = 111
	GetFileFailed                  = 112
	NotAllowUploadThisTypeFile     = 113
	UploadFileToMinioError         = 114
	BusinessHasExist               = 115
)

const (
	SuccessMsg                        = "Success"
	DBErrorMsg                        = "Internal error"
	NoApplicationNotAllowOperationMsg = "No application not allow operation"
	MissingMandatoryItemMsg           = "Missing mandatory item"
	EmailCodeErrorMsg                 = "Email code error"
	BuildJwtErrorMsg                  = "Build jwt token error"
	SendEmailErrorMsg                 = "Send email error"
	HasBindTotpErrorMsg               = "Has bind totp error"
	ValidateTOTPErrorMsg              = "Validate TOTP error"
	ParseMultipartFormErrorMsg        = "Parse multipart form error"
	GetFileFailedMsg                  = "Get file failed"
	NotAllowUploadThisTypeFileMsg     = "Not allow upload this type file"
	BusinessHasExistMsg               = "Business has exist"
)
