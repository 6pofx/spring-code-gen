package gen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// RedisConfigTemplate Redis 配置类模板
const RedisConfigTemplate = `package {{.ConfigPackageFull}};

{{if eq .JsonLib "jackson"}}import com.fasterxml.jackson.annotation.JsonAutoDetect;
import com.fasterxml.jackson.annotation.PropertyAccessor;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.jsontype.impl.LaissezFaireSubTypeValidator;
{{end}}{{if eq .JsonLib "fastjson2"}}import com.alibaba.fastjson2.support.spring.data.redis.GenericFastJsonRedisSerializer;
{{end}}{{if eq .JsonLib "gson"}}import com.google.gson.Gson;
import org.springframework.data.redis.serializer.GenericToStringSerializer;
{{end}}import org.springframework.cache.annotation.EnableCaching;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.data.redis.connection.RedisConnectionFactory;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.data.redis.serializer.StringRedisSerializer;
{{if eq .JsonLib "jackson"}}import org.springframework.data.redis.serializer.Jackson2JsonRedisSerializer;
{{end}}

@Configuration
@EnableCaching
public class RedisConfig {

    @Bean
    public RedisTemplate<String, Object> redisTemplate(RedisConnectionFactory factory) {
        RedisTemplate<String, Object> template = new RedisTemplate<>();
        template.setConnectionFactory(factory);

        // key 序列化
        StringRedisSerializer stringSerializer = new StringRedisSerializer();
        template.setKeySerializer(stringSerializer);
        template.setHashKeySerializer(stringSerializer);

        // value 序列化
{{if eq .JsonLib "jackson"}}        Jackson2JsonRedisSerializer<Object> jacksonSerializer = jacksonSerializer();
{{else if eq .JsonLib "fastjson2"}}        GenericFastJsonRedisSerializer fastJsonSerializer = new GenericFastJsonRedisSerializer();
{{else}}        Gson gson = new Gson();
        GenericToStringSerializer<Object> gsonSerializer = new GenericToStringSerializer<>(Object.class);
{{end}}
{{if eq .JsonLib "jackson"}}        template.setValueSerializer(jacksonSerializer);
        template.setHashValueSerializer(jacksonSerializer);
{{else if eq .JsonLib "fastjson2"}}        template.setValueSerializer(fastJsonSerializer);
        template.setHashValueSerializer(fastJsonSerializer);
{{else}}        template.setValueSerializer(gsonSerializer);
        template.setHashValueSerializer(gsonSerializer);
{{end}}
        template.afterPropertiesSet();
        return template;
    }
{{if eq .JsonLib "jackson"}}
    private Jackson2JsonRedisSerializer<Object> jacksonSerializer() {
        Jackson2JsonRedisSerializer<Object> serializer = new Jackson2JsonRedisSerializer<>(Object.class);
        ObjectMapper mapper = new ObjectMapper();
        mapper.setVisibility(PropertyAccessor.ALL, JsonAutoDetect.Visibility.ANY);
        mapper.activateDefaultTyping(LaissezFaireSubTypeValidator.instance, ObjectMapper.DefaultTyping.NON_FINAL);
        serializer.setObjectMapper(mapper);
        return serializer;
    }
{{end}}
}
`

// GenerateRedisConfig 生成 Redis 配置类
func GenerateRedisConfig(td *TemplateData) (string, error) {
	tmpl, err := template.New("redisConfig").Parse(RedisConfigTemplate)
	if err != nil {
		return "", fmt.Errorf("解析 RedisConfig 模板失败: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, td); err != nil {
		return "", fmt.Errorf("执行 RedisConfig 模板失败: %w", err)
	}
	return formatJavaCode(buf.String()), nil
}

// WriteRedisConfigFile 写入 RedisConfig.java
func WriteRedisConfigFile(td *TemplateData, content string, overwritePolicy string) (string, error) {
	dir := fmt.Sprintf("%s/%s", td.OutputDir, strings.ReplaceAll(td.ConfigPackageFull, ".", "/"))
	filePath := filepath.Join(dir, "RedisConfig.java")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败 %s: %w", dir, err)
	}
	return writeFile(filePath, content, overwritePolicy)
}
