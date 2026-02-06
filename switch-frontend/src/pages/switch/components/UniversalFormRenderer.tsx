import React, { useState, useMemo, useRef, useImperativeHandle, forwardRef } from 'react';
import {
  Input,
  InputNumber,
  Select,
  Switch,
  DatePicker,
  Button,
  Card,
  Tag,
  Divider
} from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { useIntl } from '@umijs/max';

// JSON Schema v7 类型定义
interface JSONSchema7 {
  $schema?: string;
  $id?: string;
  title?: string;
  description?: string;
  type?: 'null' | 'boolean' | 'object' | 'array' | 'number' | 'string' | 'integer';
  enum?: any[];
  const?: any;
  multipleOf?: number;
  maximum?: number;
  exclusiveMaximum?: number;
  minimum?: number;
  exclusiveMinimum?: number;
  maxLength?: number;
  minLength?: number;
  pattern?: string;
  items?: JSONSchema7 | JSONSchema7[];
  maxItems?: number;
  minItems?: number;
  uniqueItems?: boolean;
  contains?: JSONSchema7;
  maxProperties?: number;
  minProperties?: number;
  required?: string[];
  properties?: { [key: string]: JSONSchema7 };
  patternProperties?: { [key: string]: JSONSchema7 };
  additionalProperties?: boolean | JSONSchema7;
  dependencies?: { [key: string]: JSONSchema7 | string[] };
  propertyNames?: JSONSchema7;
  if?: JSONSchema7;
  then?: JSONSchema7;
  else?: JSONSchema7;
  allOf?: JSONSchema7[];
  anyOf?: JSONSchema7[];
  oneOf?: JSONSchema7[];
  not?: JSONSchema7;
  format?: string;
  default?: any;
  examples?: any[];
  // 自定义扩展属性
  'x-component'?: string;
  'x-component-props'?: Record<string, any>;
  'x-display'?: 'visible' | 'hidden' | 'none';
  'x-decorator'?: string;
}

interface UniversalFormRendererProps {
  input: JSONSchema7;
  value?: any;
  onChange?: (value: any) => void;
  readonly?: boolean;
}

const isValidSchema = (input: any): boolean => {
  return input && typeof input === 'object' && (
    input.type || input.properties || input.$schema || input.items || input.enum
  );
};

const UniversalFormRenderer: React.FC<UniversalFormRendererProps> = ({
  input,
  value,
  onChange,
  readonly = false
}) => {
  const intl = useIntl();
  const schema = useMemo(() => {
    if (isValidSchema(input)) {
      return input;
    }

    console.warn('UniversalFormRenderer: input is not a valid JSON Schema');
    return { type: 'string', title: 'Value' } as JSONSchema7;
  }, [input]);

  const getDefaultValue = (schema: JSONSchema7) => {
    if (schema.default !== undefined) {
      return schema.default;
    }

    switch (schema.type) {
      case 'object':
        return {};
      case 'array':
        return [];
      case 'string':
        return '';
      case 'number':
      case 'integer':
        return 0;
      case 'boolean':
        return false;
      case 'null':
        return null;
      default:
        return undefined;
    }
  };

  const formData = useMemo(() => {
    if (value === undefined) {
      return getDefaultValue(schema);
    }

    if (typeof value === 'string') {
      if (value === '') {
        return getDefaultValue(schema);
      }

      if (schema.type === 'string' || schema.type === 'number' || schema.type === 'boolean') {
        return value;
      }

      try {
        return JSON.parse(value);
      } catch (error) {
        console.warn('Failed to parse JSON string:', value, error);
        return getDefaultValue(schema);
      }
    }

    return value;
  }, [value, schema]);

  const handleChange = (path: string[], newValue: any) => {
    if (path.length === 0) {
      if (schema.type === 'string' || schema.type === 'number' || schema.type === 'boolean') {
        onChange?.(newValue);
      } else {
        const result = typeof value === 'string' ? JSON.stringify(newValue) : newValue;
        onChange?.(result);
      }
      return;
    }

    const updatedData = JSON.parse(JSON.stringify(formData));
    let current = updatedData;

    // 根据路径找要修改的值
    for (let i = 0; i < path.length - 1; i++) {
      const key = path[i];
      if (!current[key]) {
        const nextKey = path[i + 1];
        //user.items.0.name处理这样的路径
        current[key] = /^\d+$/.test(nextKey) ? [] : {};
      }
      current = current[key];
    }

    const finalKey = path[path.length - 1];
    current[finalKey] = newValue;

    const result = typeof value === 'string' ? JSON.stringify(updatedData) : updatedData;
    onChange?.(result);
  };

  // 字段的处理(囊括json schema7的字段)
  const renderField = (fieldSchema: JSONSchema7, fieldPath: string[], fieldValue: any): React.ReactNode => {
    const fieldName = fieldPath[fieldPath.length - 1];
    const { type, title, description, enum: enumValues, format } = fieldSchema;

    // 处理自定义组件
    if (fieldSchema['x-component']) {
      return renderCustomComponent(fieldSchema, fieldPath, fieldValue);
    }

    switch (type) {
      case 'string':
        // 处理类型是字符串带枚举项的情况
        if (enumValues) {
          return (
            <Select
              placeholder={`${intl.formatMessage({id: 'pages.switch.formRenderer.pleaseSelect'})}${title || fieldName}`}
              value={fieldValue}
              onChange={(val) => handleChange(fieldPath, val)}
              disabled={readonly}
              style={{ width: '100%' }}
            >
              {enumValues.map((option, index) => (
                <Select.Option key={index} value={option}>
                  {option}
                </Select.Option>
              ))}
            </Select>
          );
        }
        // data 年月日
        if (format === 'date') {
          return (
            <DatePicker
              value={fieldValue ? dayjs(fieldValue) : null}
              onChange={(date) => handleChange(fieldPath, date?.format('YYYY-MM-DD'))}
              disabled={readonly}
              style={{ width: '100%' }}
            />
          );
        }

        // data 时分秒
        if (format === 'time') {
          return (
            <DatePicker
              picker="time"
              value={fieldValue ? dayjs(fieldValue, 'HH:mm:ss') : null}
              onChange={(time) => handleChange(fieldPath, time?.format('HH:mm:ss'))}
              disabled={readonly}
              style={{ width: '100%' }}
            />
          );
        }

        // data 年月日-时分秒
        if (format === 'date-time') {
          return (
            <DatePicker
              showTime
              value={fieldValue ? dayjs(fieldValue) : null}
              onChange={(datetime) => handleChange(fieldPath, datetime?.toISOString())}
              disabled={readonly}
              style={{ width: '100%' }}
            />
          );
        }

        return (
          <Input
            placeholder={description || `${intl.formatMessage({id: 'pages.switch.formRenderer.pleaseInput'})}${title || fieldName || ""}`}
            value={fieldValue}
            onChange={(e) => handleChange(fieldPath, e.target.value)}
            disabled={readonly}
          />
        );

      case 'number':
      case 'integer':
        return (
          <InputNumber
            placeholder={description || `${intl.formatMessage({id: 'pages.switch.formRenderer.pleaseInput'})}${title || fieldName}`}
            value={fieldValue}
            onChange={(val) => handleChange(fieldPath, val)}
            disabled={readonly}
            style={{ width: '100%' }}
            min={fieldSchema.minimum}
            max={fieldSchema.maximum}
            step={fieldSchema.multipleOf || (type === 'integer' ? 1 : 0.1)}
          />
        );

      case 'boolean':
        return (
          <Switch
            checked={fieldValue}
            onChange={(checked) => handleChange(fieldPath, checked)}
            disabled={readonly}
          />
        );

      case 'array':
        return renderArrayField(fieldSchema, fieldPath, fieldValue);

      case 'object':
        return renderObjectField(fieldSchema, fieldPath, fieldValue);

      default:
        return <span>{intl.formatMessage({id: 'pages.switch.formRenderer.unsupportedType'})}: {type}</span>;
    }
  };

  const renderCustomComponent = (fieldSchema: JSONSchema7, fieldPath: string[], fieldValue: any) => {
    const component = fieldSchema['x-component'];
    const props = fieldSchema['x-component-props'] || {};

    switch (component) {
      case 'select':
        return (
          <Select
            placeholder={`${intl.formatMessage({id: 'pages.switch.formRenderer.pleaseSelect'})}${fieldSchema.title}`}
            value={fieldValue}
            onChange={(val) => handleChange(fieldPath, val)}
            disabled={readonly}
            style={{ width: '100%' }}
            {...props}
          />
        );
      default:
        return renderField({ ...fieldSchema, 'x-component': undefined }, fieldPath, fieldValue);
    }
  };


  const renderArrayField = (fieldSchema: JSONSchema7, fieldPath: string[], fieldValue: any) => {
    const items = fieldSchema.items as JSONSchema7;

    // 确保 fieldValue 是数组
    const arrayValue = Array.isArray(fieldValue) ? fieldValue : [];

    // 处理字符串数组的特殊情况
    if (items?.type === 'string') {
      if (items.enum) {
        // 多选枚举
        return (
          <Select
            mode="multiple"
            placeholder={`${intl.formatMessage({id: 'pages.switch.formRenderer.pleaseSelect'})}${fieldSchema.title}`}
            value={arrayValue}
            onChange={(val) => handleChange(fieldPath, val)}
            disabled={readonly}
            style={{ width: '100%' }}
          >
            {items.enum.map((option, index) => (
              <Select.Option key={index} value={option}>
                {option}
              </Select.Option>
            ))}
          </Select>
        );
      } else {
        // 普通字符串数组，使用标签输入模式
        return (
          <div style={{ border: '1px solid #d9d9d9', borderRadius: 6, padding: 12 }}>
            <div style={{ marginBottom: 8, display: 'flex', alignItems: 'baseline', gap: 8 }}>
              <span style={{ fontWeight: 500, lineHeight: '22px' }}>
                {fieldSchema.title || intl.formatMessage({id: 'pages.switch.formRenderer.stringList'})}
              </span>
              {fieldSchema.description && (
                <span style={{ fontSize: 12, color: '#8c8c8c', lineHeight: '22px' }}>
                  {fieldSchema.description}
                </span>
              )}
            </div>

            {arrayValue.map((item: string, index: number) => (
              <div key={index} style={{ display: 'flex', gap: 8, marginBottom: 8, alignItems: 'center' }}>
                <Input
                  value={item}
                  onChange={(e) => {
                    const newArray = [...arrayValue];
                    newArray[index] = e.target.value;
                    handleChange(fieldPath, newArray);
                  }}
                  placeholder={intl.formatMessage({id: 'pages.switch.formRenderer.inputItem'}, {index: index + 1})}
                  disabled={readonly}
                />
                {!readonly && (
                  <Button
                    type="text"
                    size="small"
                    danger
                    icon={<DeleteOutlined />}
                    onClick={() => {
                      const newArray = [...arrayValue];
                      newArray.splice(index, 1);
                      handleChange(fieldPath, newArray);
                    }}
                  />
                )}
              </div>
            ))}

            {!readonly && (
              <Button
                type="dashed"
                size="small"
                icon={<PlusOutlined />}
                onClick={() => {
                  const newArray = [...arrayValue, ''];
                  handleChange(fieldPath, newArray);
                }}
                style={{ width: '100%' }}
              >
                {intl.formatMessage({id: 'pages.switch.formRenderer.add'})}
              </Button>
            )}
          </div>
        );
      }
    }

    return (
      <div style={{ border: '1px solid #d9d9d9', borderRadius: 6, padding: 16 }}>
        <div style={{ marginBottom: 12, display: 'flex', alignItems: 'baseline', gap: 8 }}>
          <span style={{ fontWeight: 500, color: '#262626', lineHeight: '22px' }}>
            {fieldSchema.title || intl.formatMessage({id: 'pages.switch.formRenderer.arrayItems'})}
          </span>
          {fieldSchema.description && (
            <span style={{ fontSize: 12, color: '#8c8c8c', lineHeight: '22px' }}>
              {fieldSchema.description}
            </span>
          )}
        </div>

        {arrayValue.map((item: any, index: number) => (
          <Card
            key={index}
            size="small"
            title={`${intl.formatMessage({id: 'pages.switch.formRenderer.condition'})} ${index + 1}`}
            style={{ marginBottom: 8 }}
            extra={
              !readonly && (
                <Button
                  type="text"
                  size="small"
                  danger
                  icon={<DeleteOutlined />}
                  onClick={() => {
                    const newArray = [...arrayValue];
                    newArray.splice(index, 1);
                    handleChange(fieldPath, newArray);
                  }}
                />
              )
            }
          >
            {renderField(items!, [...fieldPath, index.toString()], item)}
          </Card>
        ))}

        {arrayValue.length === 0 && (
          <div style={{
            textAlign: 'center',
            padding: '20px',
            color: '#8c8c8c',
            background: '#fafafa',
            borderRadius: 4,
            marginBottom: 12
          }}>
            {intl.formatMessage({id: 'pages.switch.formRenderer.noItems'})}
          </div>
        )}

        {!readonly && (
          <Button
            type="dashed"
            icon={<PlusOutlined />}
            onClick={() => {
              const newItem = items?.type === 'object' ? {} : '';
              const newArray = [...arrayValue, newItem];
              handleChange(fieldPath, newArray);
            }}
            style={{ width: '100%' }}
          >
            {intl.formatMessage({id: 'pages.switch.formRenderer.addItem'})}
          </Button>
        )}
      </div>
    );
  };

  const renderObjectField = (fieldSchema: JSONSchema7, fieldPath: string[], fieldValue: any = {}) => {
    const { properties = {}, required = [] } = fieldSchema;

    return (
      <div style={{ border: '1px solid #d9d9d9', borderRadius: 6, padding: 16 }}>
        {Object.entries(properties).map(([propName, propSchema]) => (
          <div key={propName} style={{ marginBottom: 16 }}>
            <div style={{ marginBottom: 8, display: 'flex', alignItems: 'baseline', gap: 8 }}>
              <span style={{ fontWeight: 500, lineHeight: '22px' }}>
                {propSchema.title || propName}
              </span>
              {propSchema.description && (
                <span style={{ fontSize: 12, color: '#666', lineHeight: '22px' }}>
                  {propSchema.description}
                </span>
              )}
              {required.includes(propName) && (
                <Tag color="red" style={{ lineHeight: '20px' }}>{intl.formatMessage({id: 'pages.switch.formRenderer.required'})}</Tag>
              )}
            </div>
            {renderField(propSchema, [...fieldPath, propName], fieldValue[propName])}
          </div>
        ))}
      </div>
    );
  };

  return (
    <div style={{ width: '100%' }}>
      {schema.title && (
        <div style={{ marginBottom: 16, fontSize: 16, fontWeight: 600 }}>
          {schema.title}
        </div>
      )}
      {renderField(schema, [], formData)}
    </div>
  );
};

export default UniversalFormRenderer;
