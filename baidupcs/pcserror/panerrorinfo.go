package pcserror

import (
	"fmt"
)

type (
	// PanErrorInfo 网盘网页的api错误
	PanErrorInfo struct {
		Operation string
		ErrType   ErrType
		Err       error
		ErrNo     int `json:"errno"`
	}
)

// NewPanErrorInfo 提供operation操作名称, 返回 *PanErrorInfo
func NewPanErrorInfo(operation string) *PanErrorInfo {
	return &PanErrorInfo{
		Operation: operation,
		ErrType:   ErrorTypeNoError,
	}
}

// SetJSONError 设置JSON错误
func (pane *PanErrorInfo) SetJSONError(err error) {
	pane.ErrType = ErrTypeJSONParseError
	pane.Err = err
}

// SetNetError 设置网络错误
func (pane *PanErrorInfo) SetNetError(err error) {
	pane.ErrType = ErrTypeNetError
	pane.Err = err
}

// SetRemoteError 设置远端服务器错误
func (pane *PanErrorInfo) SetRemoteError() {
	pane.ErrType = ErrTypeRemoteError
}

// GetOperation 获取操作
func (pane *PanErrorInfo) GetOperation() string {
	return pane.Operation
}

// GetErrType 获取错误类型
func (pane *PanErrorInfo) GetErrType() ErrType {
	return pane.ErrType
}

// GetRemoteErrCode 获取远端服务器错误代码
func (pane *PanErrorInfo) GetRemoteErrCode() int {
	return pane.ErrNo
}

// GetRemoteErrMsg 获取远端服务器错误消息
func (pane *PanErrorInfo) GetRemoteErrMsg() string {
	return FindPanErr(pane.ErrNo)
}

// GetError 获取原始错误
func (pane *PanErrorInfo) GetError() error {
	return pane.Err
}

func (pane *PanErrorInfo) Error() string {
	if pane.Operation == "" {
		if pane.Err != nil {
			return pane.Err.Error()
		}
		return StrSuccess
	}

	switch pane.ErrType {
	case ErrTypeInternalError:
		return fmt.Sprintf("%s: %s, %s", pane.Operation, StrInternalError, pane.Err)
	case ErrTypeJSONParseError:
		return fmt.Sprintf("%s: %s, %s", pane.Operation, StrJSONParseError, pane.Err)
	case ErrTypeNetError:
		return fmt.Sprintf("%s: %s, %s", pane.Operation, StrNetError, pane.Err)
	case ErrTypeRemoteError:
		if pane.ErrNo == 0 {
			return fmt.Sprintf("%s: %s", pane.Operation, StrSuccess)
		}

		errmsg := FindPanErr(pane.ErrNo)
		return fmt.Sprintf("%s: encountered %s, code: %d, message: %s", pane.Operation, StrRemoteError, pane.ErrNo, errmsg)
	case ErrTypeOthers:
		if pane.Err == nil {
			return fmt.Sprintf("%s: %s", pane.Operation, StrSuccess)
		}

		return fmt.Sprintf("%s, encountered error: %s", pane.Operation, pane.Err)
	default:
		panic("panerrorinfo: unknown ErrType")
	}
}

// FindPanErr 根据 ErrNo, 解析网盘错误信息
func FindPanErr(errno int) (errmsg string) {
	switch errno {
	case 0:
		return StrSuccess
	case -1:
		return "Sharing has been disabled because prohibited content was shared. Previously shared files are unaffected."
	case -2:
		return "User does not exist, please refresh and try again"
	case -3:
		return "File does not exist, please refresh and try again"
	case -4:
		return "Login information is invalid, please log in again"
	case -5:
		return "host_key and user_key are invalid"
	case -6:
		return "Please log in again"
	case -7:
		return "This share has been deleted or cancelled"
	case -8:
		return "A file with the same name already exists"
	case -9:
		return "File does not exist"
	case -10:
		return "Share-link limit reached (100000), cannot share again"
	case -11:
		return "Verification cookie is invalid"
	case -12:
		return "Incorrect extraction password"
	case -14:
		return "SMS sharing is limited to 20 per day, please try again tomorrow"
	case -15:
		return "Email sharing is limited to 20 per day, please try again tomorrow"
	case -16:
		return "This file is restricted from sharing"
	case -17:
		return "File sharing exceeds limit"
	case -19:
		return "Captcha input is required"
	case -21:
		return "Share was cancelled or share info is invalid"
	case -30:
		return "File already exists"
	case -31:
		return "Failed to save file"
	case -33:
		return "Only up to 999 items are supported per operation"
	case -62:
		return "Captcha may be required"
	case -70:
		return "Shared file contains or may contain a virus; please share a different file"
	case 2:
		return "Please try again later, or change the save path"
	case 3:
		return "Not logged in or account is invalid"
	case 4:
		return "Storage seems unavailable, please try again later"
	case 105:
		return "Invalid link, file not found. Please open the correct share link"
	case 108:
		return "Filename contains sensitive words, please modify it"
	case 110:
		return "Share count exceeded limit. Check shared links in \"My Shares\""
	case 112:
		return "Page has expired, refresh and try again"
	case 113:
		return "Signature error"
	case 114:
		return "Current task does not exist, save failed"
	case 115:
		return "This file is forbidden to share"
	case 132:
		return "Your account may have security risks; complete security verification first"
	case 9019:
		return "accessToken is not set or has expired, use the setastoken command"
	default:
		return "Unknown error"
	}
}
