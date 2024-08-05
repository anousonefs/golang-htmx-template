package user

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/anousonefs/golang-htmx-template/internal/activity"
	"github.com/anousonefs/golang-htmx-template/internal/config"
	"github.com/anousonefs/golang-htmx-template/internal/middleware"
	"github.com/anousonefs/golang-htmx-template/internal/user/views"
	"github.com/anousonefs/golang-htmx-template/internal/utils"

	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
)

type handler struct {
	user Service
}

func NewHandler(e *echo.Echo, user Service, cfg config.Config) *handler {
	return &handler{
		user,
	}
}

func (h *handler) Install(e *echo.Echo, cfg config.Config) {
	api := e.Group("/api/v1/users", middleware.Auth(cfg)...)
	api.POST("", h.createUser)
	api.GET("", h.listUsers)
	api.GET("/:id", h.getUser)
	api.POST("/upload", h.uploadAvatar)

	roles := e.Group("/api/v1/roles", middleware.Auth(cfg)...)
	roles.GET("", h.listRoles)
	roles.POST("/:id/permissions", h.createPermission)
	roles.GET("/:id/permissions", h.listPermission)

	permissions := e.Group("/api/v1/permissions", middleware.Auth(cfg)...)
	permissions.GET("", h.listAllPermission)

	e.GET("/users", h.usersPage)
}

func (h *handler) usersPage(c echo.Context) error {
	if err := views.UserPage().Render(c.Request().Context(), c.Response().Writer); err != nil {
		return err
	}
	return nil
}

func (h *handler) listUsers(c echo.Context) error {
	ctx := c.Request().Context()
	filter := FilterUser{
		ID:       c.QueryParam("id"),
		Username: c.QueryParam("username"),
		Phone:    c.QueryParam("phone"),
	}
	res, err := h.user.ListUsers(ctx, filter)
	if err != nil {
		hs := HttpStatusPbFromRPC(GRPCStatusFromErr(err))
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(int(hs.Error.Code), b)
	}
	return c.JSON(http.StatusOK, res)
}

func (h *handler) createUser(c echo.Context) error {
	var req User
	var act activity.Activity
	if err := c.Bind(&req); err != nil {
		logrus.Errorf("bind: %v\n", err)
		hs := HttpStatusPbFromRPC(StatusBindingFailure)
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(int(hs.Error.Code), b)
	}
	if err := req.Validate(); err != nil {
		hs := HttpStatusPbFromRPC(StatusBadRequest)
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(int(hs.Error.Code), b)
	}
	ctx := c.Request().Context()
	act.CreatedBy = middleware.UserClaimFromContext(ctx).ID
	act.DepartmentID = middleware.UserClaimFromContext(ctx).DepartmentID
	req.CreatedBy = middleware.UserClaimFromContext(ctx).ID
	if err := h.user.CreateUser(ctx, req, act); err != nil {
		hs := HttpStatusPbFromRPC(GRPCStatusFromErr(err))
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(int(hs.Error.Code), b)
	}
	return c.JSON(http.StatusCreated, echo.Map{"message": "created"})
}

func (h *handler) getUser(c echo.Context) error {
	id := c.Param("id")
	ctx := c.Request().Context()
	res, err := h.user.GetUser(ctx, FilterUser{ID: id})
	if err != nil {
		hs := HttpStatusPbFromRPC(GRPCStatusFromErr(err))
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(int(hs.Error.Code), b)
	}
	return c.JSON(http.StatusOK, res)
}

func (r *handler) uploadAvatar(c echo.Context) error {
	ctx := c.Request().Context()
	userID := middleware.UserClaimFromContext(ctx).ID
	if userID == "" {
		hs := HttpStatusPbFromRPC(StatusPermissionDenied)
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(int(hs.Error.Code), b)
	}
	file, err := c.FormFile("file")
	if err != nil {
		hs := HttpStatusPbFromRPC(StatusBindingFailure)
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(int(hs.Error.Code), b)
	}

	bucketName := os.Getenv("MINIO_BUCKET")

	buffer, err := file.Open()
	if err != nil {
		hs := HttpStatusPbFromRPC(StatusBindingFailure)
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(int(hs.Error.Code), b)
	}
	defer buffer.Close()

	minioClient, err := MinioConnection()
	if err != nil {
		fmt.Printf("connection error: %v\n", err)
		hs := HttpStatusPbFromRPC(StatusInternalServerError)
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(int(hs.Error.Code), b)
	}

	objectName := file.Filename
	fileBuffer := buffer
	contentType := file.Header["Content-Type"][0]
	fileSize := file.Size

	info, err := minioClient.PutObject(ctx, bucketName, "/avatar/"+objectName, fileBuffer, fileSize, minio.PutObjectOptions{ContentType: contentType, PartSize: partSize})
	if err != nil {
		hs := HttpStatusPbFromRPC(StatusInternalServerError)
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(int(hs.Error.Code), b)
	}

	log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)
	return c.JSON(http.StatusOK, echo.Map{"message": "uploaded"})
}

func MinioConnection() (*minio.Client, error) {
	ctx := context.Background()
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESSKEY")
	secretAccessKey := os.Getenv("MINIO_SECRETKEY")
	bucketName := os.Getenv("MINIO_BUCKET")

	fmt.Printf("env: %v, %v, %v, %v\n", endpoint, accessKeyID, secretAccessKey, bucketName)

	useSSL := false
	// Initialize minio client object.
	minioClient, errInit := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if errInit != nil {
		log.Fatalln(errInit)
	}

	location := "us-east-1"

	err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}
	return minioClient, errInit
}

func (r handler) listRoles(c echo.Context) error {
	ctx := c.Request().Context()
	res, err := r.user.ListRoles(ctx)
	if err != nil {
		hs := HttpStatusPbFromRPC(GRPCStatusFromErr(err))
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(int(hs.Error.Code), b)
	}
	return c.JSON(http.StatusOK, res)
}

func (r handler) listAllPermission(c echo.Context) error {
	ctx := c.Request().Context()
	res, err := r.user.ListAllPermissions(ctx)
	if err != nil {
		hs := HttpStatusPbFromRPC(GRPCStatusFromErr(err))
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(int(hs.Error.Code), b)
	}
	return c.JSON(http.StatusOK, res)
}

func (r handler) listPermission(c echo.Context) error {
	id := c.Param("id")
	ctx := c.Request().Context()
	res, err := r.user.ListPermissions(ctx, id)
	if err != nil {
		hs := HttpStatusPbFromRPC(GRPCStatusFromErr(err))
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(int(hs.Error.Code), b)
	}
	return c.JSON(http.StatusOK, echo.Map{"permissions": res})
}

func (r handler) createPermission(c echo.Context) error {
	var req Permission
	var err error
	if err = c.Bind(&req); err != nil {
		logrus.Errorf("bind: %v\n", err)
		hs := HttpStatusPbFromRPC(StatusBindingFailure)
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(int(hs.Error.Code), b)
	}
	if req.RoleID == "" {
		hs := HttpStatusPbFromRPC(StatusBindingFailure)
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(int(hs.Error.Code), b)
	}
	ctx := c.Request().Context()
	res, err := r.user.CreatePermission(ctx, req)
	if err != nil {
		hs := HttpStatusPbFromRPC(GRPCStatusFromErr(err))
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(int(hs.Error.Code), b)
	}
	return c.JSON(http.StatusOK, res)
}

var (
	PermissionColumn = [5]string{"create", "update", "list", "delete", "get"}
)

func (r handler) permissionKey(vals []string) bool {
	if len(vals) > 0 {
		for i := 0; i < len(vals); i++ {
			ok, _ := utils.InArray(vals[i], PermissionColumn)
			if !ok {
				return false
			}
		}
	}
	return true
}
