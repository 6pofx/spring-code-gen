package gen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// ServiceImplTemplateMybatisPlus MyBatis-Plus ServiceImpl 模板
const ServiceImplTemplateMybatisPlus = `package {{.ServiceImplPkgFull}};

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import {{.EntityPackageFull}}.{{.EntityName}};
import {{.MapperPackageFull}}.{{.EntityName}}Mapper;
import {{.ServicePackageFull}}.{{.EntityName}}Service;
import org.springframework.stereotype.Service;

/**
 * {{.TableComment}} 服务实现
 */
@Service
public class {{.EntityName}}ServiceImpl extends ServiceImpl<{{.EntityName}}Mapper, {{.EntityName}}> implements {{.EntityName}}Service {

}
`

// ServiceImplTemplateMyBatis MyBatis ServiceImpl 模板
const ServiceImplTemplateMyBatis = `package {{.ServiceImplPkgFull}};

import {{.EntityPackageFull}}.{{.EntityName}};
import {{.MapperPackageFull}}.{{.EntityName}}Mapper;
import {{.ServicePackageFull}}.{{.EntityName}}Service;
{{if eq .InjectType "Resource"}}{{if eq .Namespace "jakarta"}}import jakarta.annotation.Resource;{{else}}import javax.annotation.Resource;{{end}}{{else}}import org.springframework.beans.factory.annotation.Autowired;{{end}}
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

/**
 * {{.TableComment}} 服务实现
 */
@Service
public class {{.EntityName}}ServiceImpl implements {{.EntityName}}Service {

    {{.InjectAnnotation}}
    private {{.EntityName}}Mapper {{.EntityVar}}Mapper;

    @Override
    @Transactional(rollbackFor = Exception.class)
    public boolean add({{.EntityName}} entity) {
        return {{.EntityVar}}Mapper.insert(entity) > 0;
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public boolean update({{.EntityName}} entity) {
        return {{.EntityVar}}Mapper.updateById(entity) > 0;
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public boolean deleteById({{.PrimaryKey.JavaType}} id) {
        return {{.EntityVar}}Mapper.deleteById(id) > 0;
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public boolean deleteBatch(List<{{.PrimaryKey.JavaType}}> ids) {
        return {{.EntityVar}}Mapper.deleteBatch(ids) > 0;
    }

    @Override
    public {{.EntityName}} getById({{.PrimaryKey.JavaType}} id) {
        return {{.EntityVar}}Mapper.selectById(id);
    }

    @Override
    public List<{{.EntityName}}> list() {
        return {{.EntityVar}}Mapper.selectList();
    }

    @Override
    public List<{{.EntityName}}> page(int page, int size) {
        return {{.EntityVar}}Mapper.selectPage((page - 1) * size, size);
    }
}
`

// ServiceImplTemplateJPA JPA ServiceImpl 模板
const ServiceImplTemplateJPA = `package {{.ServiceImplPkgFull}};

import {{.EntityPackageFull}}.{{.EntityName}};
import {{.MapperPackageFull}}.{{.EntityName}}Repository;
import {{.ServicePackageFull}}.{{.EntityName}}Service;
{{if eq .InjectType "Resource"}}{{if eq .Namespace "jakarta"}}import jakarta.annotation.Resource;{{else}}import javax.annotation.Resource;{{end}}{{else}}import org.springframework.beans.factory.annotation.Autowired;{{end}}
import org.springframework.stereotype.Service;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;

import java.util.List;

/**
 * {{.TableComment}} 服务实现
 */
@Service
public class {{.EntityName}}ServiceImpl implements {{.EntityName}}Service {

    {{.InjectAnnotation}}
    private {{.EntityName}}Repository {{.EntityVar}}Repository;

    @Override
    public {{.EntityName}} add({{.EntityName}} entity) {
        return {{.EntityVar}}Repository.save(entity);
    }

    @Override
    public {{.EntityName}} update({{.EntityName}} entity) {
        return {{.EntityVar}}Repository.save(entity);
    }

    @Override
    public void deleteById({{.PrimaryKey.JavaType}} id) {
        {{.EntityVar}}Repository.deleteById(id);
    }

    @Override
    public void deleteBatch(List<{{.PrimaryKey.JavaType}}> ids) {
        {{.EntityVar}}Repository.deleteAllById(ids);
    }

    @Override
    public {{.EntityName}} getById({{.PrimaryKey.JavaType}} id) {
        return {{.EntityVar}}Repository.findById(id).orElse(null);
    }

    @Override
    public List<{{.EntityName}}> list() {
        return {{.EntityVar}}Repository.findAll();
    }

    @Override
    public List<{{.EntityName}}> page(int page, int size) {
        Page<{{.EntityName}}> p = {{.EntityVar}}Repository.findAll(PageRequest.of(page - 1, size));
        return p.getContent();
    }
}
`

func getServiceImplTemplate(orm string) string {
	switch orm {
	case "mybatis-plus":
		return ServiceImplTemplateMybatisPlus
	case "mybatis":
		return ServiceImplTemplateMyBatis
	case "jpa":
		return ServiceImplTemplateJPA
	default:
		return ServiceImplTemplateMybatisPlus
	}
}

// GenerateServiceImpl 生成 ServiceImpl 文件
func GenerateServiceImpl(td *TemplateData) (string, error) {
	tmplText := getServiceImplTemplate(td.Orm)
	tmpl, err := template.New("serviceImpl").Parse(tmplText)
	if err != nil {
		return "", fmt.Errorf("解析 ServiceImpl 模板失败: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, td); err != nil {
		return "", fmt.Errorf("执行 ServiceImpl 模板失败: %w", err)
	}

	return formatJavaCode(buf.String()), nil
}

// WriteServiceImplFile 写入 ServiceImpl 文件
func WriteServiceImplFile(td *TemplateData, content string, overwritePolicy string) (string, error) {
	dir := td.GetServiceImplDir()
	fileName := fmt.Sprintf("%sServiceImpl.java", td.EntityName)
	filePath := filepath.Join(dir, fileName)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败 %s: %w", dir, err)
	}

	return writeFile(filePath, content, overwritePolicy)
}
