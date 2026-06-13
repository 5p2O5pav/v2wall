package master

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/yourorg/v2wall/internal/config"
	"github.com/yourorg/v2wall/internal/db"
	"golang.org/x/crypto/bcrypt"
)

// AdminUser 管理员账户结构（存储在 BadgerDB）
type AdminUser struct {
	Username string `json:"username"`
	Password string `json:"password"` // bcrypt 哈希
	Role     string `json:"role"`     // "admin"
}

// 是否存在任何用户
func hasAnyUser(bdb *badger.DB) (bool, error) {
	var found bool
	err := bdb.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("user:")
		it := txn.NewIterator(opts)
		defer it.Close()
		it.Rewind()
		if it.Valid() {
			found = true
		}
		return nil
	})
	return found, err
}

// 保存用户
func saveUser(bdb *badger.DB, user *AdminUser) error {
	key := db.UserKey(user.Username)
	val, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return bdb.Update(func(txn *badger.Txn) error {
		return txn.Set(key, val)
	})
}

// 获取用户
func getUser(bdb *badger.DB, username string) (*AdminUser, error) {
	var user AdminUser
	err := bdb.View(func(txn *badger.Txn) error {
		key := db.UserKey(username)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &user)
		})
	})
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// handleInit 初始化管理员（仅无用户时可用）
func handleInit(bdb *badger.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否已有用户
		has, err := hasAnyUser(bdb)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		if has {
			c.JSON(http.StatusForbidden, gin.H{"error": "already initialized"})
			return
		}

		// 验证临时 Token
		token := strings.TrimSpace(c.GetHeader("Authorization"))
		token = strings.TrimPrefix(token, "Bearer ")
		if token != cfg.Master.InitToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid init token"})
			return
		}

		// 解析请求体
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 创建用户
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "hash failed"})
			return
		}
		user := &AdminUser{
			Username: req.Username,
			Password: string(hash),
			Role:     "admin",
		}
		if err := saveUser(bdb, user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "save user failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "initialized"})
	}
}

// handleLogin 管理员登录，返回 JWT
func handleLogin(bdb *badger.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := getUser(bdb, req.Username)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		// 生成 JWT
		claims := jwt.MapClaims{
			"username": user.Username,
			"role":     user.Role,
			"exp":      time.Now().Add(24 * time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(cfg.Master.JWTSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": tokenString})
	}
}

// JWTAuthMiddleware 验证 JWT 的中间件
func JWTAuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}
		c.Set("username", claims["username"])
		c.Next()
	}
}

// SyncTokenMiddleware 验证同步 Token 的中间件
// 支持 Authorization: Bearer <token> 或查询参数 ?token=
func SyncTokenMiddleware(validToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先从 Authorization 头获取
		authHeader := c.GetHeader("Authorization")
		token := ""
		if authHeader != "" {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
		// 其次从查询参数获取
		if token == "" {
			token = c.Query("token")
		}
		if token != validToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid sync token"})
			return
		}
		c.Next()
	}
}
