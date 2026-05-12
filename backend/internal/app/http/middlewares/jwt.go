package middlewares

import (
	"cinema/internal/models"
	"cinema/internal/repository"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuthMiddleware(secretKey string, reps ...repository.Repository) gin.HandlerFunc {
	var rep repository.Repository
	if len(reps) > 0 {
		rep = reps[0]
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token is missing"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token format"})
			return
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
				return
			}
		}

		userID, ok := getUintClaim(claims, "user_id")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id claim"})
			return
		}

		roleID, ok := getUintClaim(claims, "role_id")
		if !ok {
			roleID = 0
		}

		status, _ := claims["status"].(string)
		username, _ := claims["username"].(string)
		roleName, _ := claims["role_name"].(string)
		tokenType, _ := claims["type"].(string)
		if tokenType != "access" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "access token required"})
			return
		}

		if rep != nil {
			user, err := new(models.User).FindByID(rep, int(userID))
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token user"})
				return
			}
			status = user.Status
			roleID = user.RoleID
			var role models.Role
			if err := rep.First(&role, user.RoleID).Error; err == nil {
				roleName = role.Name
			}
		}

		c.Set("user_id", userID)
		c.Set("username", username)
		c.Set("role_id", roleID)
		c.Set("role_name", roleName)
		c.Set("status", status)

		c.Next()
	}
}

func RequireActiveStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		status, _ := c.Get("status")
		if statusStr, ok := status.(string); !ok || statusStr != "active" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user is not active"})
			return
		}
		c.Next()
	}
}

func RequireRoles(allowedRoleIDs ...uint) gin.HandlerFunc {
	allowed := make(map[uint]struct{}, len(allowedRoleIDs))
	for _, r := range allowedRoleIDs {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "missing role"})
			return
		}
		roleID, ok := roleVal.(uint)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid role"})
			return
		}
		if _, ok := allowed[roleID]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

func RequireRoleNames(allowedRoleNames ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedRoleNames))
	for _, name := range allowedRoleNames {
		normalized := strings.ToLower(strings.TrimSpace(name))
		if normalized != "" {
			allowed[normalized] = struct{}{}
		}
	}
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role_name")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "missing role"})
			return
		}
		roleName, ok := roleVal.(string)
		if !ok || strings.TrimSpace(roleName) == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid role"})
			return
		}
		if _, ok := allowed[strings.ToLower(strings.TrimSpace(roleName))]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

func getUintClaim(claims jwt.MapClaims, key string) (uint, bool) {
	val, ok := claims[key]
	if !ok {
		return 0, false
	}
	floatVal, ok := val.(float64)
	if !ok {
		return 0, false
	}
	if floatVal < 0 {
		return 0, false
	}
	return uint(floatVal), true
}
