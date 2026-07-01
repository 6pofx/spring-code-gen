package gen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// ControllerTemplate 通用 Controller 模板
const ControllerTemplate = `package {{.ControllerPkgFull}};

import {{.EntityPackageFull}}.{{.EntityName}};
import {{.ServicePackageFull}}.{{.EntityName}}Service;
import {{.BasePackage}}.{{.ResponseClass}};
{{if eq .Orm "mybatis-plus"}}import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
{{end}}{{if eq .Orm "mybatis"}}import java.util.List;
{{end}}{{if eq .Orm "jpa"}}import org.springframework.data.domain.Page;
import java.util.List;
{{end}}{{if eq .InjectType "Resource"}}{{if eq .Namespace "jakarta"}}import jakarta.annotation.Resource;{{else}}import javax.annotation.Resource;{{end}}{{else}}import org.springframework.beans.factory.annotation.Autowired;{{end}}
import org.springframework.web.bind.annotation.*;
{{if eq .Swagger "springdoc"}}import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
{{else}}import io.swagger.annotations.Api;
import io.swagger.annotations.ApiOperation;
{{end}}
/**
 * {{.TableComment}} 控制器
 */
{{if eq .Swagger "springdoc"}}@Tag(name = "{{.TableComment}}")
{{else}}@Api(tags = "{{.TableComment}}")
{{end}}@RestController
@RequestMapping("/{{.EntityPath}}")
public class {{.EntityName}}Controller {

    @{{.InjectType}}
    private {{.EntityName}}Service {{.EntityVar}}Service;

{{if eq .Swagger "springdoc"}}    @Operation(summary = "分页查询")
{{else}}    @ApiOperation("分页查询")
{{end}}    @GetMapping("/page")
    public {{.ResponseClass}}<Page<{{.EntityName}}>> page(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "10") int size) {
{{if eq .Orm "mybatis-plus"}}        Page<{{.EntityName}}> p = (Page<{{.EntityName}}>) {{.EntityVar}}Service.page(new Page<>(page, size));
{{else}}        Page<{{.EntityName}}> p = (Page<{{.EntityName}}>) {{.EntityVar}}Service.page(page, size);
{{end}}        return {{.ResponseClass}}.ok(p);
    }

{{if eq .Swagger "springdoc"}}    @Operation(summary = "全部列表")
{{else}}    @ApiOperation("全部列表")
{{end}}    @GetMapping("/list")
    public {{.ResponseClass}}<java.util.List<{{.EntityName}}>> list() {
        java.util.List<{{.EntityName}}> list = {{.EntityVar}}Service.list();
        return {{.ResponseClass}}.ok(list);
    }

{{if eq .Swagger "springdoc"}}    @Operation(summary = "根据ID查询")
{{else}}    @ApiOperation("根据ID查询")
{{end}}    @GetMapping("/{id}")
    public {{.ResponseClass}}<{{.EntityName}}> get(@PathVariable {{.PrimaryKey.JavaType}} id) {
        {{.EntityName}} entity = {{.EntityVar}}Service.getById(id);
        return {{.ResponseClass}}.ok(entity);
    }

{{if eq .Swagger "springdoc"}}    @Operation(summary = "新增")
{{else}}    @ApiOperation("新增")
{{end}}    @PostMapping
    public {{.ResponseClass}}<Void> add(@RequestBody {{.EntityName}} entity) {
{{if eq .Orm "mybatis-plus"}}        {{.EntityVar}}Service.save(entity);
{{else}}        {{.EntityVar}}Service.add(entity);
{{end}}        return {{.ResponseClass}}.ok();
    }

{{if eq .Swagger "springdoc"}}    @Operation(summary = "修改")
{{else}}    @ApiOperation("修改")
{{end}}    @PutMapping
    public {{.ResponseClass}}<Void> update(@RequestBody {{.EntityName}} entity) {
{{if eq .Orm "mybatis-plus"}}        {{.EntityVar}}Service.updateById(entity);
{{else}}        {{.EntityVar}}Service.update(entity);
{{end}}        return {{.ResponseClass}}.ok();
    }

{{if eq .Swagger "springdoc"}}    @Operation(summary = "删除")
{{else}}    @ApiOperation("删除")
{{end}}    @DeleteMapping("/{id}")
    public {{.ResponseClass}}<Void> delete(@PathVariable {{.PrimaryKey.JavaType}} id) {
{{if eq .Orm "mybatis-plus"}}        {{.EntityVar}}Service.removeById(id);
{{else}}        {{.EntityVar}}Service.deleteById(id);
{{end}}        return {{.ResponseClass}}.ok();
    }

{{if eq .Swagger "springdoc"}}    @Operation(summary = "批量删除")
{{else}}    @ApiOperation("批量删除")
{{end}}    @DeleteMapping("/batch")
    public {{.ResponseClass}}<Void> deleteBatch(@RequestParam String ids) {
        String[] idArr = ids.split(",");
{{if eq .PrimaryKey.JavaType "Long"}}        java.util.List<Long> idList = new java.util.ArrayList<>();
{{else}}        java.util.List<{{.PrimaryKey.JavaType}}> idList = new java.util.ArrayList<>();
{{end}}        for (String id : idArr) {
{{if eq .PrimaryKey.JavaType "Long"}}            idList.add(Long.parseLong(id.trim()));
{{else}}            idList.add({{.PrimaryKey.JavaType}}.parse{{.PrimaryKey.JavaType}}(id.trim()));
{{end}}        }
{{if eq .Orm "mybatis-plus"}}        {{.EntityVar}}Service.removeByIds(idList);
{{else}}        {{.EntityVar}}Service.deleteBatch(idList);
{{end}}        return {{.ResponseClass}}.ok();
    }
}
`

// GenerateController 生成 Controller 文件
func GenerateController(td *TemplateData) (string, error) {
	tmpl, err := template.New("controller").Parse(ControllerTemplate)
	if err != nil {
		return "", fmt.Errorf("解析 Controller 模板失败: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, td); err != nil {
		return "", fmt.Errorf("执行 Controller 模板失败: %w", err)
	}

	return formatJavaCode(buf.String()), nil
}

// WriteControllerFile 写入 Controller 文件
func WriteControllerFile(td *TemplateData, content string, overwritePolicy string) (string, error) {
	dir := td.GetControllerDir()
	fileName := fmt.Sprintf("%sController.java", td.EntityName)
	filePath := filepath.Join(dir, fileName)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败 %s: %w", dir, err)
	}

	return writeFile(filePath, content, overwritePolicy)
}
