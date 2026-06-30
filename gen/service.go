package gen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// ServiceTemplateMybatisPlus MyBatis-Plus Service 接口模板
const ServiceTemplateMybatisPlus = `package {{.ServicePackageFull}};

import com.baomidou.mybatisplus.extension.service.IService;
import {{.EntityPackageFull}}.{{.EntityName}};

/**
 * {{.TableComment}} 服务接口
 */
public interface {{.EntityName}}Service extends IService<{{.EntityName}}> {

}
`

// ServiceTemplateMyBatis MyBatis Service 接口模板
const ServiceTemplateMyBatis = `package {{.ServicePackageFull}};

import {{.EntityPackageFull}}.{{.EntityName}};

import java.util.List;

/**
 * {{.TableComment}} 服务接口
 */
public interface {{.EntityName}}Service {

    /**
     * 新增
     */
    boolean add({{.EntityName}} entity);

    /**
     * 修改
     */
    boolean update({{.EntityName}} entity);

    /**
     * 根据ID删除
     */
    boolean deleteById({{.PrimaryKey.JavaType}} id);

    /**
     * 批量删除
     */
    boolean deleteBatch(List<{{.PrimaryKey.JavaType}}> ids);

    /**
     * 根据ID查询
     */
    {{.EntityName}} getById({{.PrimaryKey.JavaType}} id);

    /**
     * 查询全部
     */
    List<{{.EntityName}}> list();

    /**
     * 分页查询
     */
    List<{{.EntityName}}> page(int page, int size);
}
`

// ServiceTemplateJPA JPA Service 接口模板
const ServiceTemplateJPA = `package {{.ServicePackageFull}};

import {{.EntityPackageFull}}.{{.EntityName}};

import java.util.List;

/**
 * {{.TableComment}} 服务接口
 */
public interface {{.EntityName}}Service {

    /**
     * 新增
     */
    {{.EntityName}} add({{.EntityName}} entity);

    /**
     * 修改
     */
    {{.EntityName}} update({{.EntityName}} entity);

    /**
     * 根据ID删除
     */
    void deleteById({{.PrimaryKey.JavaType}} id);

    /**
     * 批量删除
     */
    void deleteBatch(List<{{.PrimaryKey.JavaType}}> ids);

    /**
     * 根据ID查询
     */
    {{.EntityName}} getById({{.PrimaryKey.JavaType}} id);

    /**
     * 查询全部
     */
    List<{{.EntityName}}> list();

    /**
     * 分页查询
     */
    List<{{.EntityName}}> page(int page, int size);
}
`

func getServiceTemplate(orm string) string {
	switch orm {
	case "mybatis-plus":
		return ServiceTemplateMybatisPlus
	case "mybatis":
		return ServiceTemplateMyBatis
	case "jpa":
		return ServiceTemplateJPA
	default:
		return ServiceTemplateMybatisPlus
	}
}

// GenerateService 生成 Service 接口文件
func GenerateService(td *TemplateData) (string, error) {
	tmplText := getServiceTemplate(td.Orm)
	tmpl, err := template.New("service").Parse(tmplText)
	if err != nil {
		return "", fmt.Errorf("解析 Service 模板失败: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, td); err != nil {
		return "", fmt.Errorf("执行 Service 模板失败: %w", err)
	}

	return formatJavaCode(buf.String()), nil
}

// WriteServiceFile 写入 Service 接口文件
func WriteServiceFile(td *TemplateData, content string, overwritePolicy string) (string, error) {
	dir := td.GetServiceDir()
	fileName := fmt.Sprintf("%sService.java", td.EntityName)
	filePath := filepath.Join(dir, fileName)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败 %s: %w", dir, err)
	}

	return writeFile(filePath, content, overwritePolicy)
}
