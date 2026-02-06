import {ModalForm, ProFormText, ProFormTextArea} from '@ant-design/pro-components';
import {message, Button, Select, Input, Space, Card} from 'antd';
import React, {useState, useEffect} from 'react';
import {createSwitchFactor, updateSwitchFactor} from '@/services/switch-factor/api';
import {useIntl, useRequest} from '@umijs/max';
import {PlusOutlined, DeleteOutlined, ExclamationCircleOutlined} from '@ant-design/icons';
import {useErrorHandler} from '@/utils/useErrorHandler';

export type CreateFormProps = {
  isCreate: boolean,
  values?: API.SwitchFactorItem;
  reload: () => void;
  trigger: React.ReactNode;
  readOnly?: boolean;
};


// JSON Schema 编辑器
// 使用的是json schema 第七版规范
// JSON Schema 校验函数
const validateJsonSchema = (schemaStr: string, intl: any): { isValid: boolean; errors: string[] } => {
  if (!schemaStr.trim()) {
    return { isValid: false, errors: [intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.inputContent' })] };
  }

  try {
    const schema = JSON.parse(schemaStr);
    const errors: string[] = [];

    // 基本结构校验
    if (typeof schema !== 'object' || schema === null) {
      errors.push(intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.mustBeObject' }));
    }

    // 检查 type 字段
    if (schema.type && !['string', 'number', 'integer', 'boolean', 'object', 'array', 'null'].includes(schema.type)) {
      errors.push(`${intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.unsupportedType' })}: "${schema.type}"`);
    }

    // 检查 properties 字段
    if (schema.properties && typeof schema.properties !== 'object') {
      errors.push(intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.propertiesMustBeObject' }));
    }

    // 检查 required 字段
    if (schema.required) {
      if (!Array.isArray(schema.required)) {
        errors.push(intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.requiredMustBeArray' }));
      } else {
        // 检查 required 数组中的每个元素是否为字符串
        schema.required.forEach((item: any, index: number) => {
          if (typeof item !== 'string') {
            errors.push(`required[${index}] ${intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.requiredItemMustBeString' })}`);
          }
        });

        // 如果有 properties 字段，检查 required 中的属性是否在 properties 中存在
        if (schema.properties && typeof schema.properties === 'object') {
          schema.required.forEach((requiredProp: string) => {
            if (typeof requiredProp === 'string' && !schema.properties.hasOwnProperty(requiredProp)) {
              errors.push(`${intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.requiredField' })} "${requiredProp}" ${intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.requiredFieldNotInProperties' })}`);
            }
          });
        } else if (schema.required.length > 0) {
          errors.push(intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.requiredFieldMissingProperties' }));
        }
      }
    }

    // 检查 items 字段
    if (schema.items && typeof schema.items !== 'object') {
      errors.push(intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.itemsMustBeObject' }));
    }

    // 检查数值约束
    if (schema.minimum !== undefined && typeof schema.minimum !== 'number') {
      errors.push(intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.minimumMustBeNumber' }));
    }
    if (schema.maximum !== undefined && typeof schema.maximum !== 'number') {
      errors.push(intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.maximumMustBeNumber' }));
    }
    if (schema.minLength !== undefined && (typeof schema.minLength !== 'number' || schema.minLength < 0)) {
      errors.push(intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.minLengthMustBeNonNegative' }));
    }
    if (schema.maxLength !== undefined && (typeof schema.maxLength !== 'number' || schema.maxLength < 0)) {
      errors.push(intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.maxLengthMustBeNonNegative' }));
    }

    // 检查数值范围关系
    if (schema.minimum !== undefined && schema.maximum !== undefined && schema.minimum > schema.maximum) {
      errors.push(intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.minCannotExceedMax' }));
    }

    if (schema.minLength !== undefined && schema.maxLength !== undefined && schema.minLength > schema.maxLength) {
      errors.push(intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.minLengthCannotExceedMaxLength' }));
    }

    // 检查类型相关的约束
    if (schema.type === 'object') {
      if (!schema.properties && schema.additionalProperties === false) {
        errors.push(intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.objectNeedsProperties' }));
      }
    }

    if (schema.type === 'array') {
      if (!schema.items) {
        errors.push(intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.arrayNeedsItems' }));
      }
    }

    // 检查字段适用性
    if (schema.type && schema.type !== 'string' && (schema.minLength !== undefined || schema.maxLength !== undefined || schema.pattern !== undefined)) {
      errors.push(intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.lengthConstraintsForString' }));
    }

    if (schema.type && !['number', 'integer'].includes(schema.type) && (schema.minimum !== undefined || schema.maximum !== undefined || schema.multipleOf !== undefined)) {
      errors.push(intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.numericConstraintsForNumber' }));
    }

    if (schema.type && schema.type !== 'array' && (schema.minItems !== undefined || schema.maxItems !== undefined || schema.uniqueItems !== undefined)) {
      errors.push(intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.arrayConstraintsForArray' }));
    }

    return { isValid: errors.length === 0, errors };
  } catch (e) {
    return { isValid: false, errors: [intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.syntaxError' }) + ': ' + (e as Error).message] };
  }
};

const JsonSchemaEditor: React.FC<{
  value?: string;
  onChange?: (value: string) => void;
  readOnly?: boolean;
  validationErrors?: string[];
  intl: any;
}> = ({ value, onChange, readOnly = false, validationErrors = [], intl }) => {
  const [schemaData, setSchemaData] = useState<any>({});

  const typeOptions = [
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.type.string' }), value: 'string' },
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.type.number' }), value: 'number' },
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.type.integer' }), value: 'integer' },
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.type.boolean' }), value: 'boolean' },
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.type.object' }), value: 'object' },
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.type.array' }), value: 'array' },
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.type.null' }), value: 'null' }
  ];

  const formatOptions = [
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.format.dateTime' }), value: 'date-time' },
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.format.date' }), value: 'date' },
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.format.time' }), value: 'time' },
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.format.email' }), value: 'email' },
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.format.hostname' }), value: 'hostname' },
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.format.ipv4' }), value: 'ipv4' },
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.format.ipv6' }), value: 'ipv6' },
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.format.uri' }), value: 'uri' },
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.format.uriReference' }), value: 'uri-reference' },
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.format.iri' }), value: 'iri' },
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.format.iriReference' }), value: 'iri-reference' },
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.format.uuid' }), value: 'uuid' },
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.format.jsonPointer' }), value: 'json-pointer' },
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.format.relativeJsonPointer' }), value: 'relative-json-pointer' },
    { label: intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.format.regex' }), value: 'regex' }
  ];

  // 递归处理 JSON Schema，自动推断缺失的 type
  const processSchemaData = (data: any): any => {
    if (typeof data !== 'object' || data === null) {
      return data;
    }

    const processed = { ...data };

    if ((processed.format || processed.pattern) && !processed.type) {
      processed.type = 'string';
    }

    //以下内容需要递归处理
    ['anyOf', 'oneOf', 'allOf'].forEach(key => {
      if (Array.isArray(processed[key])) {
        processed[key] = processed[key].map((item: any) => processSchemaData(item));
      }
    });

    if (processed.items) {
      processed.items = processSchemaData(processed.items);
    }

    if (processed.properties && typeof processed.properties === 'object') {
      const newProperties: any = {};
      Object.keys(processed.properties).forEach(key => {
        newProperties[key] = processSchemaData(processed.properties[key]);
      });
      processed.properties = newProperties;
    }

    return processed;
  };

  useEffect(() => {
    if (value) {
      try {
        const parsed = JSON.parse(value);
        setSchemaData(parsed);
      } catch (e) {
        console.error('Invalid JSON Schema:', e);
      }
    }
  }, [value]);


  const updateSchema = (newData: any) => {
    if (readOnly) return;
    setSchemaData(newData);
    onChange?.(JSON.stringify(newData, null, 2));
  };

  // 根据type类型获取可用的属性
  const getAvailableProperties = (type: string) => {
    const commonProps = ['description', 'title', 'default', 'examples', 'anyOf', 'oneOf', 'allOf'];

    switch (type) {
      case 'string':
        return [...commonProps, 'minLength', 'maxLength', 'pattern', 'format', 'enum'];
      case 'number':
      case 'integer':
        return [...commonProps, 'minimum', 'maximum', 'multipleOf', 'enum'];
      case 'array':
        return [...commonProps, 'items', 'minItems', 'maxItems', 'uniqueItems'];
      case 'object':
        return [...commonProps, 'properties', 'required', 'additionalProperties', 'minProperties', 'maxProperties'];
      case 'boolean':
        return [...commonProps];
      default:
        return commonProps;
    }
  };

  const addProperty = (propName: string) => {
    const newData = { ...schemaData };

    switch (propName) {
      case 'properties':
        newData.properties = newData.properties || {};
        break;
      case 'required':
        newData.required = newData.required || [];
        break;
      case 'enum':
      case 'examples':
        newData[propName] = [];
        break;
      case 'anyOf':
      case 'oneOf':
      case 'allOf':
        newData[propName] = [{ type: 'string' }];
        break;
      case 'items':
        newData.items = { type: 'string' };
        break;
      case 'additionalProperties':
      case 'uniqueItems':
        newData[propName] = false;
        break;
      default:
        newData[propName] = '';
    }

    updateSchema(newData);
  };

  const removeProperty = (propName: string) => {
    const newData = { ...schemaData };
    delete newData[propName];
    updateSchema(newData);
  };

  // 每次更新schema的时候。同时校验schema && 更新schema的值
  const updateProperty = (propName: string, propValue: any) => {
    const newData = { ...schemaData };
    newData[propName] = propValue;

    if (propName === 'type') {
      const availableProps = getAvailableProperties(propValue);
      Object.keys(newData).forEach(key => {
        if (key !== 'type' && !availableProps.includes(key)) {
          delete newData[key];
        }
      });
    }

    updateSchema(newData);
  };

  const renderPropertyEditor = (propName: string, propValue: any) => {
    switch (propName) {
      case '$schema':
        return (
          <Input
            value={propValue}
            style={{ width: '100%' }}
            disabled={true}
            placeholder={intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.schemaVersionReadonly' })}
          />
        );

      case 'type':
        return (
          <Select
            value={propValue}
            style={{ width: '100%' }}
            options={typeOptions}
            onChange={(val) => updateProperty(propName, val)}
            placeholder={intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.selectDataType' })}
            disabled={readOnly}
          />
        );

      case 'format':
        return (
          <Select
            value={propValue}
            style={{ width: '100%' }}
            options={formatOptions}
            onChange={(val) => updateProperty(propName, val)}
            placeholder={intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.selectFormat' })}
            disabled={readOnly}
          />
        );

      case 'required':
        return (
          <Select
            mode="tags"
            value={propValue}
            style={{ width: '100%' }}
            placeholder={intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.selectRequiredFields' })}
            onChange={(val) => updateProperty(propName, val)}
            disabled={readOnly}
          />
        );

      case 'enum':
      case 'examples':
        return (
          <Select
            mode="tags"
            value={propValue}
            style={{ width: '100%' }}
            placeholder={propName === 'enum' ? intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.inputEnumValue' }) : intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.inputExampleValue' })}
            onChange={(val) => updateProperty(propName, val)}
            disabled={readOnly}
          />
        );

      case 'items':
        return (
          <Card size="small" style={{ marginTop: '8px' }}>
            <div style={{ marginBottom: '8px', fontWeight: 'bold' }}>{intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.arrayItemType' })}</div>
            <JsonSchemaEditor
              value={JSON.stringify(propValue, null, 2)}
              onChange={(val) => {
                try {
                  updateProperty(propName, JSON.parse(val));
                } catch (e) {
                }
              }}
              readOnly={readOnly}
              intl={intl}
            />
          </Card>
        );

      case 'anyOf':
      case 'oneOf':
      case 'allOf':
        return (
          <Card size="small" style={{ marginTop: '8px' }}>
            {Array.isArray(propValue) && propValue.map((item, index) => (
              <div key={index} style={{ marginBottom: '12px', padding: '8px', border: '1px solid #f0f0f0', borderRadius: '4px' }}>
                <Space style={{ width: '100%', justifyContent: 'space-between', marginBottom: '8px' }}>
                  <strong>{intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.option' })} {index + 1}:</strong>
                  {!readOnly && (
                    <Button
                      type="text"
                      size="small"
                      icon={<DeleteOutlined />}
                      onClick={() => {
                        const newArray = [...propValue];
                        newArray.splice(index, 1);
                        updateProperty(propName, newArray);
                      }}
                      danger
                    />
                  )}
                </Space>
                <JsonSchemaEditor
                  value={JSON.stringify(item, null, 2)}
                  onChange={(val) => {
                    try {
                      const newArray = [...propValue];
                      newArray[index] = JSON.parse(val);
                      updateProperty(propName, newArray);
                    } catch (e) {
                    }
                  }}
                  readOnly={readOnly}
                  intl={intl}
                />
              </div>
            ))}
            {!readOnly && (
              <Button
                type="dashed"
                icon={<PlusOutlined />}
                onClick={() => {
                  const newArray = Array.isArray(propValue) ? [...propValue] : [];
                  newArray.push({ type: 'string' });
                  updateProperty(propName, newArray);
                }}
                style={{ width: '100%' }}
              >
                {intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.addOption' })}
              </Button>
            )}
          </Card>
        );

      case 'properties':
        return (
          <Card size="small" style={{ marginTop: '8px' }}>
            <div style={{ marginBottom: '8px', fontWeight: 'bold' }}>{intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.objectProperties' })}</div>
            {Object.keys(propValue || {}).map(key => {
              return (
                <div key={key} style={{ marginBottom: '12px', padding: '8px', border: '1px solid #f0f0f0', borderRadius: '4px' }}>
                  <Space style={{ width: '100%', justifyContent: 'space-between', marginBottom: '8px' }}>
                    <strong>
                      {key}:
                    </strong>
                    {!readOnly && (
                      <Button
                        type="text"
                        size="small"
                        icon={<DeleteOutlined />}
                        onClick={() => {
                          const newProps = { ...propValue };
                          delete newProps[key];
                          updateProperty(propName, newProps);
                        }}
                        danger
                      />
                    )}
                  </Space>
                  <JsonSchemaEditor
                    value={JSON.stringify(propValue[key], null, 2)}
                    onChange={(val) => {
                      try {
                        const newProps = { ...propValue };
                        newProps[key] = JSON.parse(val);
                        updateProperty(propName, newProps);
                      } catch (e) {
                      }
                    }}
                    readOnly={readOnly}
                    intl={intl}
                  />
                </div>
              );
            })}
            <Input
              placeholder={intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.inputNewPropertyName' })}
              onPressEnter={(e) => {
                const target = e.target as HTMLInputElement;
                const newKey = target.value.trim();
                if (newKey) {
                  const newProps = { ...propValue };
                  newProps[newKey] = { type: 'string' };
                  updateProperty(propName, newProps);
                  target.value = '';
                }
              }}
              key={`property-input-${Object.keys(propValue || {}).length}`}
              disabled={readOnly}
            />
          </Card>
        );

      case 'minimum':
      case 'maximum':
      case 'minLength':
      case 'maxLength':
      case 'minItems':
      case 'maxItems':
      case 'minProperties':
      case 'maxProperties':
      case 'multipleOf':
        return (
          <Input
            type="number"
            value={propValue}
            onChange={(e) => updateProperty(propName, Number(e.target.value))}
            placeholder={`${intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.input' })}${propName}`}
            disabled={readOnly}
          />
        );

      case 'uniqueItems':
      case 'additionalProperties':
        return (
          <Select
            value={propValue}
            style={{ width: '100%' }}
            options={[
              { label: 'true', value: true },
              { label: 'false', value: false }
            ]}
            onChange={(val) => updateProperty(propName, val)}
            disabled={readOnly}
          />
        );

      default:
        return (
          <Input
            value={propValue}
            onChange={(e) => updateProperty(propName, e.target.value)}
            placeholder={`${intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.input' })}${propName}`}
            disabled={readOnly}
          />
        );
    }
  };

  const currentType = schemaData.type;
  const availableProps = currentType ? getAvailableProperties(currentType) : [];
  const unusedProps = availableProps.filter(prop => !schemaData.hasOwnProperty(prop));

  return (
    <div style={{ border: validationErrors.length > 0 ? '1px solid #ff4d4f' : '1px solid #d9d9d9', borderRadius: '6px', padding: '12px', maxHeight: '400px', overflowY: 'auto' }}>
      {validationErrors.length > 0 && (
        <div style={{
          background: 'linear-gradient(135deg, #fff2f0 0%, #ffebe6 100%)',
          border: '1px solid #ffccc7',
          borderRadius: '8px',
          padding: '16px',
          marginBottom: '16px',
          boxShadow: '0 2px 8px rgba(255, 77, 79, 0.1)'
        }}>
          <div style={{
            display: 'flex',
            alignItems: 'center',
            marginBottom: '12px',
            color: '#cf1322',
            fontWeight: 600,
            fontSize: '14px'
          }}>
            <ExclamationCircleOutlined style={{ marginRight: '8px', fontSize: '16px' }} />
            {intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.validationFailed' })}
          </div>
          <div style={{ fontSize: '13px', lineHeight: '1.6' }}>
            {validationErrors.map((error, index) => (
              <div
                key={index}
                style={{
                  display: 'flex',
                  alignItems: 'flex-start',
                  marginBottom: index < validationErrors.length - 1 ? '8px' : '0',
                  color: '#8c1f1f'
                }}
              >
                <span style={{
                  display: 'inline-block',
                  width: '6px',
                  height: '6px',
                  borderRadius: '50%',
                  backgroundColor: '#ff4d4f',
                  marginTop: '6px',
                  marginRight: '10px',
                  flexShrink: 0
                }} />
                <span>{error}</span>
              </div>
            ))}
          </div>
        </div>
      )}
      <Space direction="vertical" style={{ width: '100%' }}>
        <div>
          <strong>type:</strong>
          {renderPropertyEditor('type', schemaData.type)}
        </div>

        {Object.keys(schemaData).filter(key => key !== 'type').map(propName => {
          return (
            <div key={propName}>
              <Space style={{ width: '100%', justifyContent: 'space-between', marginBottom: '4px' }}>
                <strong>
                  {propName}:
                </strong>
                {propName !== '$schema' && !readOnly && (
                  <Button
                    type="text"
                    size="small"
                    icon={<DeleteOutlined />}
                    onClick={() => removeProperty(propName)}
                    danger
                  />
                )}
              </Space>
              {renderPropertyEditor(propName, schemaData[propName])}
            </div>
          );
        })}

        {unusedProps.length > 0 && !readOnly && (
          <Select
            placeholder={intl.formatMessage({ id: 'pages.switch-factor.jsonSchema.selectPropertyToAdd' })}
            style={{ width: '100%' }}
            onSelect={(propName: string) => addProperty(propName)}
            options={unusedProps.map(prop => ({
              label: prop,
              value: prop
            }))}
          />
        )}
      </Space>
    </div>
  );
};

const CreateUpdateForm: React.FC<CreateFormProps> = (props) => {
  const {values, trigger, isCreate, readOnly = false} = props;
  const intl = useIntl();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [jsonSchemaValue, setJsonSchemaValue] = useState<string>('');
  const [schemaValidationErrors, setSchemaValidationErrors] = useState<string[]>([]);

  const [messageApi, contextHolder] = message.useMessage();
  const { handleError } = useErrorHandler();

  useEffect(() => {
    if (isModalOpen) {
      if (values?.jsonSchema) {
        setJsonSchemaValue(values.jsonSchema);
      } else {
        setJsonSchemaValue('');
      }
      setSchemaValidationErrors([]);
    }
  }, [values, isModalOpen]);

  const {run, loading} = useRequest(
    (params: Partial<API.SwitchFactorItem>) => {
      return isCreate ? createSwitchFactor(params) : updateSwitchFactor(params);
    },
    {
      manual: true,
      onSuccess: () => {
        const successMsg = isCreate ? intl.formatMessage({ id: 'pages.switch-factor.form.createSuccess' }) : intl.formatMessage({ id: 'pages.switch-factor.form.updateSuccess' });
        messageApi.success(successMsg);
        setIsModalOpen(false);
        props.reload();
      },
      onError: (error) => {
        handleError(
          error,
          isCreate ? 'pages.switch-factor.form.createError' : 'pages.switch-factor.form.updateError',
          { showDetail: true }
        );
        setIsModalOpen(false);
      },
    },
  );

  const renderTrigger = () => {
    return <span onClick={() => setIsModalOpen(true)}>{trigger}</span>;
  };

  return (
    <>
      {contextHolder}
      {renderTrigger()}
      <ModalForm
        width={900}
        style={{
          paddingTop: '10px',
          marginLeft: "20px"
        }}
        labelCol={{flex: "0 0 25%"}}
        wrapperCol={{flex: "0 0 75%"}}
        title={isCreate ? intl.formatMessage({
          id: 'pages.switch-factor.searchTable.create',
        }) : readOnly ? intl.formatMessage({
          id: 'pages.switch-factor.searchTable.view',
        }) : intl.formatMessage({
          id: 'pages.switch-factor.searchTable.editChild',
        })}
        open={isModalOpen}
        onOpenChange={setIsModalOpen}
        initialValues={values}
        submitter={readOnly ? false : {
          searchConfig: {
            resetText: intl.formatMessage({ id: 'pages.switch-factor.form.resetText' }),
            submitText: intl.formatMessage({ id: 'pages.switch-factor.form.submitText' }),
          },
          submitButtonProps: {
            type: 'primary',
            style: {
              marginLeft: "10px"
            },
            loading: loading,
          },
          resetButtonProps: {
            onClick: () => {
              // 重置应该清空当前表单数据和 JSON Schema
              setJsonSchemaValue('');
              setSchemaValidationErrors([]);
            },
          },
        }}
        modalProps={{
          styles: {
            header: {
              paddingTop: '10px',
              paddingLeft: '10px'
            },
          },
          destroyOnHidden: true
        }}
        onFinish={readOnly ? undefined : async (formValues) => {
          // 校验 JSON Schema
          const validation = validateJsonSchema(jsonSchemaValue, intl);
          if (!validation.isValid) {
            setSchemaValidationErrors(validation.errors);
            return false;
          }

          // 清空校验错误
          setSchemaValidationErrors([]);

          const params = {...values, ...formValues, jsonSchema: jsonSchemaValue};
          await run(params);
        }}
      >
        <ProFormText
          name="factor"
          label={
            <>
              {intl.formatMessage({ id: 'pages.switch-factor.form.factorName' })}
              <span style={{color: '#999', fontSize: '12px', marginLeft: '8px'}}>
                {intl.formatMessage({ id: 'pages.switch-factor.form.factorNameReadonly' })}
                </span>
            </>
          }
          formItemProps={{
            style: {
              marginTop: "15px",
            },
            labelCol: {
              style: {
                paddingBottom: "10px",
              }
            },
          }}
          fieldProps={{
            style: {
              width: "95%"
            },
            disabled: !isCreate || readOnly
          }}

          rules={[{required: true, message: intl.formatMessage({ id: 'pages.switch-factor.form.factorNamePlaceholder' })}]}
        />
        <ProFormText
          name="name"
          label={intl.formatMessage({ id: 'pages.switch-factor.form.factorAlias' })}
          formItemProps={{
            style: {
              marginTop: "15px",
            },
            labelCol: {
              style: {
                paddingBottom: "10px",
              }
            },
          }}
          fieldProps={{
            style: {
              width: "95%"
            },
            disabled: !isCreate || readOnly
          }}

          rules={[{required: true, message: intl.formatMessage({ id: 'pages.switch-factor.form.factorAliasPlaceholder' })}]}
        />
        <ProFormTextArea
          name="description"
          label={intl.formatMessage({ id: 'pages.switch-factor.form.description' })}
          formItemProps={{
            style: {
              marginTop: "20px",
              marginBottom: "60px"
            },
            labelCol: {
              style: {
                paddingBottom: "10px",
              }
            },
          }}
          fieldProps={{
            style: {
              width: "95%"
            },
            disabled: readOnly
          }}
          rules={[{required: true, message: intl.formatMessage({ id: 'pages.switch-factor.form.descriptionPlaceholder' })}]}
        />

        <div style={{ marginTop: "20px", marginBottom: "60px" }}>
          <div style={{
            marginBottom: "8px",
            color: "rgba(0, 0, 0, 0.88)",
            fontSize: "14px"
          }}>
            <label>{intl.formatMessage({ id: 'pages.switch-factor.form.jsonSchema' })}</label>
          </div>
          <JsonSchemaEditor
            value={jsonSchemaValue}
            onChange={readOnly ? undefined : (value) => {
              setJsonSchemaValue(value);
              // 实时校验
              if (value.trim()) {
                const validation = validateJsonSchema(value, intl);
                setSchemaValidationErrors(validation.errors);
              } else {
                setSchemaValidationErrors([]);
              }
            }}
            readOnly={readOnly}
            validationErrors={schemaValidationErrors}
            intl={intl}
          />
        </div>
      </ModalForm>
    </>
  );
};

export default CreateUpdateForm;
