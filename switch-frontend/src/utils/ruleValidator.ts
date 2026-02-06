// 类型定义
interface SwitchFactorItem {
  id: number;
  name?: string;
  factor?: string;
  jsonSchema?: string;
}

// 校验结果接口
export interface ValidationResult {
  isValid: boolean;
  errors: ValidationError[];
}

export interface ValidationError {
  path: string;
  messageKey: string; // 国际化 key
  messageParams?: Record<string, any>; // 国际化参数
  factorId?: number;
  factorName?: string;
}

// 根据JSON Schema校验单个config值
const validateConfigBySchema = (
  config: any,
  jsonSchema: any,
  factorId: number,
  factorName: string,
  path: string = ''
): ValidationError[] => {
  const errors: ValidationError[] = [];

  if (!jsonSchema || typeof jsonSchema !== 'object') {
    return errors;
  }

  const { type, required, properties } = jsonSchema;

  // 校验必填字段
  if (required && Array.isArray(required)) {
    required.forEach((field: string) => {
      if (!config || config[field] === undefined || config[field] === null || config[field] === '') {
        errors.push({
          path: path ? `${path}.${field}` : field,
          messageKey: 'pages.ruleValidator.fieldRequired',
          messageParams: { field },
          factorId,
          factorName
        });
      }
    });
  }

  // 校验类型
  if (type && config !== undefined && config !== null) {
    switch (type) {
      case 'object':
        if (typeof config !== 'object' || Array.isArray(config)) {
          const actualType = Array.isArray(config) ? 'array' : typeof config;
          errors.push({
            path,
            messageKey: 'pages.ruleValidator.typeExpectedObject',
            messageParams: { actualType },
            factorId,
            factorName
          });
        } else if (properties) {
          // 递归校验对象属性
          Object.keys(properties).forEach(key => {
            const subErrors = validateConfigBySchema(
              config[key],
              properties[key],
              factorId,
              factorName,
              path ? `${path}.${key}` : key
            );
            errors.push(...subErrors);
          });
        }
        break;
      case 'array':
        if (!Array.isArray(config)) {
          errors.push({
            path,
            messageKey: 'pages.ruleValidator.typeExpectedArray',
            messageParams: { actualType: typeof config },
            factorId,
            factorName
          });
        }
        break;
      case 'string':
        if (typeof config !== 'string') {
          errors.push({
            path,
            messageKey: 'pages.ruleValidator.typeExpectedString',
            messageParams: { actualType: typeof config },
            factorId,
            factorName
          });
        }
        break;
      case 'number':
        if (typeof config !== 'number') {
          errors.push({
            path,
            messageKey: 'pages.ruleValidator.typeExpectedNumber',
            messageParams: { actualType: typeof config },
            factorId,
            factorName
          });
        }
        break;
      case 'boolean':
        if (typeof config !== 'boolean') {
          errors.push({
            path,
            messageKey: 'pages.ruleValidator.typeExpectedBoolean',
            messageParams: { actualType: typeof config },
            factorId,
            factorName
          });
        }
        break;
    }
  }

  return errors;
};

// 递归校验规则树
const validateRuleNode = (
  node: any,
  factorOptions: SwitchFactorItem[],
  path: string = 'root'
): ValidationError[] => {
  const errors: ValidationError[] = [];

  if (!node) {
    return errors;
  }

  // 判断节点类型
  const isGroupNode = node.nodeType && (node.nodeType === 'AND' || node.nodeType === 'OR');
  const isFactorNode = !node.nodeType;

  // 如果没有nodeType，必须是因子节点
  if (isFactorNode) {
    // 因子节点的基本校验
    if (!node.id || node.id === 0) {
      errors.push({
        path,
        messageKey: 'pages.ruleValidator.factorIdRequired',
        factorId: node.id,
        factorName: node.factor
      });
    }

    if (!node.factor || node.factor === '') {
      errors.push({
        path,
        messageKey: 'pages.ruleValidator.factorNameRequired',
        factorId: node.id,
        factorName: node.factor
      });
    }

    if (node.config === undefined || node.config === null || node.config === '') {
      errors.push({
        path,
        messageKey: 'pages.ruleValidator.factorConfigRequired',
        factorId: node.id,
        factorName: node.factor
      });
    }
  }

  // 如果是因子节点且基本信息有效
  if (isFactorNode && node.id && node.factor) {
    const factor = factorOptions.find(f => f.id === node.id);

    if (!factor) {
      errors.push({
        path,
        messageKey: 'pages.ruleValidator.factorNotFound',
        messageParams: { id: node.id },
        factorId: node.id,
        factorName: node.factor
      });
      return errors;
    }

    // 解析JSON Schema
    let jsonSchema;
    try {
      jsonSchema = JSON.parse(factor.jsonSchema || '{}');
    } catch (error) {
      errors.push({
        path,
        messageKey: 'pages.ruleValidator.jsonSchemaInvalid',
        messageParams: { name: factor.name },
        factorId: node.id,
        factorName: node.factor
      });
      return errors;
    }

    // 校验config
    const configErrors = validateConfigBySchema(
      node.config,
      jsonSchema,
      node.id,
      factor.name || node.factor,
      `${path}.config`
    );
    errors.push(...configErrors);
  }

  // 如果是组节点
  if (isGroupNode) {
    // 组节点必须有children
    if (!node.children || !Array.isArray(node.children) || node.children.length === 0) {
      errors.push({
        path,
        messageKey: 'pages.ruleValidator.groupNodeChildrenRequired',
        messageParams: { nodeType: node.nodeType },
        factorId: node.id,
        factorName: node.factor
      });
    } else {
      // 递归校验子节点
      node.children.forEach((child: any, index: number) => {
        const childErrors = validateRuleNode(
          child,
          factorOptions,
          `${path}.children[${index}]`
        );
        errors.push(...childErrors);
      });
    }
  }

  // 如果既不是组节点也不是因子节点
  if (!isGroupNode && !isFactorNode) {
    errors.push({
      path,
      messageKey: 'pages.ruleValidator.invalidNodeType',
      factorId: node.id,
      factorName: node.factor
    });
  }

  return errors;
};

// 主校验函数
export const validateParsedRule = (
  parsedRule: any,
  factorOptions: SwitchFactorItem[]
): ValidationResult => {
  const errors = validateRuleNode(parsedRule, factorOptions);

  return {
    isValid: errors.length === 0,
    errors
  };
};

// 格式化错误信息（接收国际化函数）
export const formatValidationErrors = (
  errors: ValidationError[],
  formatMessage: (descriptor: { id: string }, values?: Record<string, any>) => string
): string => {
  if (errors.length === 0) {
    return formatMessage({ id: 'pages.ruleValidator.validationPassed' });
  }

  return errors.map(error => {
    const factorInfo = error.factorName ? `[${error.factorName}] ` : '';
    const message = formatMessage({ id: error.messageKey }, error.messageParams);
    const pathLabel = formatMessage({ id: 'pages.ruleValidator.path' });
    return `${factorInfo}${message} (${pathLabel}: ${error.path})`;
  }).join('\n');
};
