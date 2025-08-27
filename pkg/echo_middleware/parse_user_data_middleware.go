package echomiddleware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	errmsg "github.com/gocasters/rankr/pkg/err_msg"
	"github.com/gocasters/rankr/type"

	"github.com/labstack/echo/v4"
)

func ParseUserDataMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		base64Data := c.Request().Header.Get("X-User-Info")

		if base64Data == "" {
			return c.JSON(http.StatusBadRequest,
				errmsg.ErrorResponse{
					Message: errmsg.ErrGetUserInfo.Error(),
					Errors: map[string]interface{}{
						"header_data_error": errmsg.MessageMissingXUserData,
					},
				},
			)

		}

		jsonData, err := base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			return c.JSON(http.StatusBadRequest,
				errmsg.ErrorResponse{
					Message: errmsg.ErrFailedDecodeBase64.Error(),
					Errors: map[string]interface{}{
						"decode_data_error": errmsg.MessageInvalidBase64,
					},
				},
			)
		}

		var userInfo types.UserClaim
		err = json.Unmarshal(jsonData, &userInfo)
		if err != nil {
			return c.JSON(http.StatusBadRequest,
				errmsg.ErrorResponse{
					Message: errmsg.ErrFailedUnmarshalJson.Error(),
					Errors: map[string]interface{}{
						"decode_data_error": errmsg.MessageInvalidJsonFormat,
					},
				},
			)
		}

		c.Set("userInfo", &userInfo)

		return next(c)
	}
}
