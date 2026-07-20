package pomcheck

import (
	"fmt"
	"strings"
)

// RequiredDep 一个必需的 Maven 依赖
type RequiredDep struct {
	GroupID       string `json:"groupId"`
	ArtifactID    string `json:"artifactId"`
	Version       string `json:"version,omitempty"` // 推荐版本，空表示由 BOM 管理
	Scope         string `json:"scope,omitempty"`
	Optional      string `json:"optional,omitempty"`
	Reason        string `json:"reason"`         // 为什么需要这个依赖
	SpringBootBOM bool   `json:"springBootBOM"`  // true = Spring Boot BOM 已管理版本
}

// RequiredDepResult 匹配检查结果
type DepMatchResult struct {
	Required  RequiredDep `json:"required"`
	Found     bool        `json:"found"`     // 在 pom.xml 中找到了
	FoundDecl string      `json:"foundDecl"` // 找到的声明片段
	SuggestedXML string   `json:"suggestedXML"` // 建议添加的 XML
}

// depsByConfig 根据生成配置推断所需依赖
func depsByConfig(dbType, orm, swagger, springVersion string, useLombok bool, enableRedis bool, redisClient, jsonLib string) []RequiredDep {
	var deps []RequiredDep

	isSB3 := springVersion == "3.x"

	// === 基础 ===
	deps = append(deps, RequiredDep{
		GroupID:       "org.springframework.boot",
		ArtifactID:    "spring-boot-starter-web",
		Reason:        "Web 支持（Controller 必需）",
		SpringBootBOM: true,
	})
	// === 数据库驱动 ===
	switch dbType {
	case "mysql":
		deps = append(deps, RequiredDep{
			GroupID:       "com.mysql",
			ArtifactID:    "mysql-connector-j",
			Scope:         "runtime",
			Reason:        "MySQL 数据库驱动",
			SpringBootBOM: true,
		})
	case "postgresql":
		deps = append(deps, RequiredDep{
			GroupID:       "org.postgresql",
			ArtifactID:    "postgresql",
			Scope:         "runtime",
			Reason:        "PostgreSQL 数据库驱动",
			SpringBootBOM: true,
		})
	}

	// === ORM ===
	switch orm {
	case "mybatis-plus":
		artifactId := "mybatis-plus-spring-boot3-starter"
		version := "3.5.9"
		if !isSB3 {
			artifactId = "mybatis-plus-boot-starter"
			version = "3.5.9"
		}
		deps = append(deps, RequiredDep{
			GroupID:    "com.baomidou",
			ArtifactID: artifactId,
			Version:    version,
			Reason:     "MyBatis-Plus ORM 框架",
		})
	case "mybatis":
		deps = append(deps, RequiredDep{
			GroupID:       "org.mybatis.spring.boot",
			ArtifactID:    "mybatis-spring-boot-starter",
			Reason: fmt.Sprintf("MyBatis ORM 框架（%s）",
				map[bool]string{true: "3.x", false: "2.x"}[isSB3]),
		})
		// MyBatis SB3 的 artifactId 不同
		if isSB3 {
			deps = append(deps[:len(deps)-1], RequiredDep{
				GroupID:    "org.mybatis.spring.boot",
				ArtifactID: "mybatis-spring-boot-starter",
				Version:    "3.0.3",
				Reason:     "MyBatis ORM 框架（Spring Boot 3.x）",
			})
		} else {
			deps = append(deps[:len(deps)-1], RequiredDep{
				GroupID:       "org.mybatis.spring.boot",
				ArtifactID:    "mybatis-spring-boot-starter",
				Reason:        "MyBatis ORM 框架（Spring Boot 2.x）",
				SpringBootBOM: false,
				Version:       "2.3.2",
			})
		}
	case "jpa":
		deps = append(deps, RequiredDep{
			GroupID:       "org.springframework.boot",
			ArtifactID:    "spring-boot-starter-data-jpa",
			Reason:        "JPA / Hibernate ORM",
			SpringBootBOM: true,
		})
		deps = append(deps, RequiredDep{
			GroupID:       "org.springframework.boot",
			ArtifactID:    "spring-boot-starter-validation",
			Reason:        "Hibernate 6 需要 Jakarta Bean Validation（实体校验 @NotNull 等）",
			SpringBootBOM: true,
		})
	}

	// === API 文档 ===
	switch swagger {
	case "springdoc":
		artifactId := "springdoc-openapi-starter-webmvc-ui"
		version := "2.6.0"
		if !isSB3 {
			artifactId = "springdoc-openapi-ui"
			version = "1.8.0"
		}
		deps = append(deps, RequiredDep{
			GroupID:    "org.springdoc",
			ArtifactID: artifactId,
			Version:    version,
			Reason:     "SpringDoc OpenAPI 文档",
		})
	case "swagger2":
		if isSB3 {
			deps = append(deps, RequiredDep{
				GroupID:    "io.springfox",
				ArtifactID: "springfox-boot-starter",
				Version:    "3.0.0",
				Reason:     "Swagger 2 API 文档（Spring Boot 3.x）",
			})
		} else {
			deps = append(deps, RequiredDep{
				GroupID:    "io.springfox",
				ArtifactID: "springfox-swagger2",
				Version:    "2.9.2",
				Reason:     "Swagger 2 API 文档",
			})
			deps = append(deps, RequiredDep{
				GroupID:    "io.springfox",
				ArtifactID: "springfox-swagger-ui",
				Version:    "2.9.2",
				Reason:     "Swagger 2 UI",
			})
		}
	}

	// === Lombok ===
	if useLombok {
		deps = append(deps, RequiredDep{
			GroupID:    "org.projectlombok",
			ArtifactID: "lombok",
			Optional:   "true",
			Version:    "1.18.46",
			Reason:     "Lombok 代码简化",
		})
	}

	// === Redis 缓存 ===
	if enableRedis {
		deps = append(deps, RequiredDep{
			GroupID:       "org.springframework.boot",
			ArtifactID:    "spring-boot-starter-data-redis",
			Reason:        "Redis 缓存支持",
			SpringBootBOM: true,
		})
		if redisClient == "jedis" {
			deps = append(deps, RequiredDep{
				GroupID:       "redis.clients",
				ArtifactID:    "jedis",
				Reason:        "Jedis Redis 客户端",
				SpringBootBOM: true,
			})
		}
		// JSON 序列化库
		switch jsonLib {
		case "fastjson2":
			deps = append(deps, RequiredDep{
				GroupID:    "com.alibaba.fastjson2",
				ArtifactID: "fastjson2",
				Version:    "2.0.53",
				Reason:     "Fastjson2 JSON 序列化",
			})
			deps = append(deps, RequiredDep{
				GroupID:    "com.alibaba.fastjson2",
				ArtifactID: "fastjson2-extension-spring6",
				Version:    "2.0.53",
				Reason:     "Fastjson2 Spring 6 扩展",
			})
		case "gson":
			deps = append(deps, RequiredDep{
				GroupID:       "com.google.code.gson",
				ArtifactID:    "gson",
				Reason:        "Gson JSON 序列化",
				SpringBootBOM: true,
			})
		}
	}

	return deps
}

// toXML 生成建议的 XML 片段
func (r RequiredDep) toXML(indent string) string {
	xml := indent + "<dependency>\n"
	xml += indent + "\t<groupId>" + r.GroupID + "</groupId>\n"
	xml += indent + "\t<artifactId>" + r.ArtifactID + "</artifactId>\n"
	if r.Version != "" {
		xml += indent + "\t<version>" + r.Version + "</version>\n"
	}
	if r.Scope != "" {
		xml += indent + "\t<scope>" + r.Scope + "</scope>\n"
	}
	if r.Optional != "" {
		xml += indent + "\t<optional>" + r.Optional + "</optional>\n"
	}
	xml += indent + "</dependency>"
	return xml
}

// CheckRequiredDeps 根据生成配置检查 pom.xml 中是否缺少必要依赖
// 返回：已存在的依赖列表、缺失的依赖列表
func CheckRequiredDeps(dbType, orm, swagger, springVersion string, useLombok bool, existingDeps []Dependency, enableRedis bool, redisClient, jsonLib string) (found []DepMatchResult, missing []DepMatchResult) {
	required := depsByConfig(dbType, orm, swagger, springVersion, useLombok, enableRedis, redisClient, jsonLib)

	// 构建已存在的 key 集合
	existing := make(map[string]bool)
	for _, d := range existingDeps {
		key := d.GroupID + ":" + d.ArtifactID
		existing[key] = true
	}

	for _, r := range required {
		key := r.GroupID + ":" + r.ArtifactID
		result := DepMatchResult{
			Required:  r,
			Found:     existing[key],
			SuggestedXML: r.toXML(""),
		}

		if existing[key] {
			// 找到匹配的声明
			for _, d := range existingDeps {
				if d.GroupID == r.GroupID && d.ArtifactID == r.ArtifactID {
					parts := []string{d.GroupID + ":" + d.ArtifactID}
					if d.Version != "" {
						parts = append(parts, d.Version)
					}
					if d.Scope != "" {
						parts = append(parts, "scope="+d.Scope)
					}
					result.FoundDecl = strings.Join(parts, " ")
					break
				}
			}
			found = append(found, result)
		} else {
			missing = append(missing, result)
		}
	}

	return
}
