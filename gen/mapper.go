package gen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// MapperTemplateMybatisPlus MyBatis-Plus Mapper 模板
const MapperTemplateMybatisPlus = `package {{.MapperPackageFull}};

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import {{.EntityPackageFull}}.{{.EntityName}};
import org.apache.ibatis.annotations.Mapper;

/**
 * {{.TableComment}} Mapper 接口
 */
@Mapper
public interface {{.EntityName}}Mapper extends BaseMapper<{{.EntityName}}> {

}
`

// MapperTemplateMyBatis MyBatis Mapper 模板
const MapperTemplateMyBatis = `package {{.MapperPackageFull}};

import {{.EntityPackageFull}}.{{.EntityName}};
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

/**
 * {{.TableComment}} Mapper 接口
 */
@Mapper
public interface {{.EntityName}}Mapper {

    /**
     * 插入记录
     */
    int insert({{.EntityName}} entity);

    /**
     * 根据ID更新
     */
    int updateById({{.EntityName}} entity);

    /**
     * 根据ID删除
     */
    int deleteById(@Param("id") {{.PrimaryKey.JavaType}} id);

    /**
     * 批量删除
     */
    int deleteBatch(@Param("ids") List<{{.PrimaryKey.JavaType}}> ids);

    /**
     * 根据ID查询
     */
    {{.EntityName}} selectById(@Param("id") {{.PrimaryKey.JavaType}} id);

    /**
     * 查询全部列表
     */
    List<{{.EntityName}}> selectList();

    /**
     * 分页查询
     */
    List<{{.EntityName}}> selectPage(@Param("offset") int offset, @Param("limit") int limit);

    /**
     * 查询总数
     */
    long selectCount();
}
`

// MapperXMLTemplate MyBatis XML 映射文件模板
const MapperXMLTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="{{.MapperPackageFull}}.{{.EntityName}}Mapper">

    <!-- 通用映射 -->
    <resultMap id="BaseResultMap" type="{{.EntityPackageFull}}.{{.EntityName}}">
{{range .Columns}}        <result column="{{.ColName}}" property="{{.JavaField}}"{{if .IsPrimaryKey}} jdbcType="{{JdbcType .JavaType}}"{{end}}/>
{{end}}    </resultMap>

    <!-- 通用查询列 -->
    <sql id="BaseColumns">
{{range $i, $c := .Columns}}{{if $i}}, {{end}}{{$c.ColName}}{{end}}    </sql>

    <!-- 插入 -->
    <insert id="insert" useGeneratedKeys="true" keyProperty="{{.PrimaryKey.JavaField}}">
        INSERT INTO {{.TableName}} (
{{range $i, $c := .NonCommonColumns}}{{if $i}}, {{end}}{{$c.ColName}}{{end}}
        ) VALUES (
{{range $i, $c := .NonCommonColumns}}{{if $i}}, {{end}}#{{JdbcParam $c.JavaField $c.JavaType}}{{end}}
        )
    </insert>

    <!-- 根据ID更新 -->
    <update id="updateById">
        UPDATE {{.TableName}}
        <set>
{{range .NonPrimaryColumns}}            <if test="{{.JavaField}} != null">{{.ColName}} = #{{JdbcParam .JavaField .JavaType}},</if>
{{end}}        </set>
        WHERE {{.PrimaryKey.ColName}} = #{{JdbcParam .PrimaryKey.JavaField .PrimaryKey.JavaType}}
    </update>

    <!-- 根据ID删除 -->
    <delete id="deleteById">
        DELETE FROM {{.TableName}} WHERE {{.PrimaryKey.ColName}} = #{{JdbcParam .PrimaryKey.JavaField .PrimaryKey.JavaType}}
    </delete>

    <!-- 批量删除 -->
    <delete id="deleteBatch">
        DELETE FROM {{.TableName}} WHERE {{.PrimaryKey.ColName}} IN
        <foreach collection="ids" item="id" open="(" separator="," close=")">
            #{{JdbcParam "id" .PrimaryKey.JavaType}}
        </foreach>
    </delete>

    <!-- 根据ID查询 -->
    <select id="selectById" resultMap="BaseResultMap">
        SELECT <include refid="BaseColumns"/> FROM {{.TableName}} WHERE {{.PrimaryKey.ColName}} = #{{JdbcParam .PrimaryKey.JavaField .PrimaryKey.JavaType}}
    </select>

    <!-- 查询全部列表 -->
    <select id="selectList" resultMap="BaseResultMap">
        SELECT <include refid="BaseColumns"/> FROM {{.TableName}}
    </select>

    <!-- 分页查询 -->
    <select id="selectPage" resultMap="BaseResultMap">
        SELECT <include refid="BaseColumns"/> FROM {{.TableName}} LIMIT #{limit} OFFSET #{offset}
    </select>

    <!-- 查询总数 -->
    <select id="selectCount" resultType="long">
        SELECT COUNT(*) FROM {{.TableName}}
    </select>

</mapper>
`

// MapperTemplateJPA JPA Repository 模板
const MapperTemplateJPA = `package {{.MapperPackageFull}};

import {{.EntityPackageFull}}.{{.EntityName}};
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.stereotype.Repository;

/**
 * {{.TableComment}} Repository
 */
@Repository
public interface {{.EntityName}}Repository extends JpaRepository<{{.EntityName}}, {{.PrimaryKey.JavaType}}>, JpaSpecificationExecutor<{{.EntityName}}> {

}
`

// MapperXMLFuncMap 用于 XML 模板的特殊函数
var MapperXMLFuncMap = template.FuncMap{
	"JdbcType":  jdbcType,
	"JdbcParam": jdbcParam,
}

func jdbcType(javaType string) string {
	switch javaType {
	case "Integer":
		return "INTEGER"
	case "Long":
		return "BIGINT"
	case "String":
		return "VARCHAR"
	case "Boolean":
		return "BOOLEAN"
	case "java.time.LocalDate":
		return "DATE"
	case "java.time.LocalDateTime":
		return "TIMESTAMP"
	case "java.math.BigDecimal":
		return "DECIMAL"
	case "Float":
		return "REAL"
	case "Double":
		return "DOUBLE"
	case "byte[]":
		return "BLOB"
	default:
		return "VARCHAR"
	}
}

func jdbcParam(field, javaType string) string {
	return "{" + field + ",jdbcType=" + jdbcType(javaType) + "}"
}

func getMapperTemplate(orm string) string {
	switch orm {
	case "mybatis-plus":
		return MapperTemplateMybatisPlus
	case "mybatis":
		return MapperTemplateMyBatis
	case "jpa":
		return MapperTemplateJPA
	default:
		return MapperTemplateMybatisPlus
	}
}

// GenerateMapper 生成 Mapper 接口文件
func GenerateMapper(td *TemplateData) (string, error) {
	tmplText := getMapperTemplate(td.Orm)
	tmpl, err := template.New("mapper").Parse(tmplText)
	if err != nil {
		return "", fmt.Errorf("解析 Mapper 模板失败: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, td); err != nil {
		return "", fmt.Errorf("执行 Mapper 模板失败: %w", err)
	}

	return formatJavaCode(buf.String()), nil
}

// GenerateMapperXML 生成 MyBatis XML 映射文件
func GenerateMapperXML(td *TemplateData) (string, error) {
	tmpl, err := template.New("mapper-xml").Funcs(MapperXMLFuncMap).Parse(MapperXMLTemplate)
	if err != nil {
		return "", fmt.Errorf("解析 Mapper XML 模板失败: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, td); err != nil {
		return "", fmt.Errorf("执行 Mapper XML 模板失败: %w", err)
	}

	return buf.String(), nil
}

// WriteMapperFile 写入 Mapper 文件
func WriteMapperFile(td *TemplateData, content string, overwritePolicy string) (string, error) {
	var suffix string
	if td.Orm == "jpa" {
		suffix = "Repository"
	} else {
		suffix = "Mapper"
	}

	dir := td.GetMapperDir()
	fileName := fmt.Sprintf("%s%s.java", td.EntityName, suffix)
	filePath := filepath.Join(dir, fileName)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败 %s: %w", dir, err)
	}

	return writeFile(filePath, content, overwritePolicy)
}

// WriteMapperXMLFile 写入 MyBatis XML 映射文件
func WriteMapperXMLFile(td *TemplateData, content string, overwritePolicy string) (string, error) {
	dir := td.GetMapperXMLDir()
	fileName := fmt.Sprintf("%sMapper.xml", td.EntityName)
	filePath := filepath.Join(dir, fileName)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败 %s: %w", dir, err)
	}

	return writeFile(filePath, content, overwritePolicy)
}

// IsCommonField 判断是否为通用字段
func IsCommonField(colName string) bool {
	lower := strings.ToLower(colName)
	_, ok := commonFields[lower]
	return ok
}
