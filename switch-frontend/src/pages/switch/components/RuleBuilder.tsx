import React, { useState, useEffect, useRef } from 'react';
import { Button,Select} from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import UniversalFormRenderer from './UniversalFormRenderer';
import { switchFactorsLike } from '@/services/switch-factor/api';
import {useErrorHandler} from "@/utils/useErrorHandler";
import {useIntl} from "@umijs/max";


interface RuleBuilderProps {
  value?: API.RuleNode;
  onChange?: (value: API.RuleNode | undefined) => void;
}

const RuleBuilder: React.FC<RuleBuilderProps> = ({ value, onChange }) => {
  const intl = useIntl();
  const [rootRule, setRootRule] = useState<API.RuleNode | undefined>(value);
  const [factorOptions, setFactorOptions] = useState<API.SwitchFactorItem[]>();
  const { handleError } = useErrorHandler();

  // 当外部 value 变化时，更新内部状态
  useEffect(() => {
    setRootRule(value);
  }, [value]);

  // 递归解析config字符串为JSON对象
  const parseConfigToJson = (node: API.RuleNode): any => {
    const parsedNode = { ...node };

    // 如果是因子节点，解析config
    if (parsedNode.config && typeof parsedNode.config === 'string') {
      try {
        parsedNode.config = JSON.parse(parsedNode.config);
      } catch (error) {
        // 如果解析失败，保持原值
        console.warn('Config解析失败:', parsedNode.config, error);
      }
    }

    // 递归处理子节点
    if (parsedNode.children) {
      parsedNode.children = parsedNode.children.map(child => parseConfigToJson(child));
    }

    return parsedNode;
  };

  // 不管是组还是因子都会更新并通知外界他们的修改
  const updateRule = (newRule: API.RuleNode | undefined) => {
    console.log('完整的规则数据:', JSON.stringify(newRule, null, 2));
    setRootRule(newRule);
    onChange?.(newRule);
  };

  // 获取因子选项数据
  useEffect(() => {
    const fetchFactorOptions = async () => {
      try {
        console.log('开始获取因子元数据...');
        const response = await switchFactorsLike({}, {}, {});
        console.log('因子元数据接口返回数据:', response);

        if (response && Array.isArray(response)) {
          setFactorOptions(response);
        }
      } catch (error) {
        handleError(
          error,
          'pages.switch.message.fetchSwitchFactorFailed',
          { showDetail: true }
        );
      }
    };

    fetchFactorOptions();
  }, []);

  const createInitialGroup = () => {
    const newGroup: API.RuleNode = {
      id: 0,
      nodeType: 'AND',
      children: []
    };
    updateRule(newGroup);
  };
  // 区分默认状态跟已添加规则的状态
  if (!rootRule) {
    return (
      <div style={{
        background: '#fafbfc',
        border: '1px solid #e1e4e8',
        borderRadius: '16px',
        padding: '24px'
      }}>
        <div style={{
          fontSize: '16px',
          fontWeight: '600',
          color: '#24292f',
          marginBottom: '20px',
          display: 'flex',
          alignItems: 'center',
          gap: '8px'
        }}>
          <div style={{
            width: '4px',
            height: '16px',
            background: '#1890ff',
            borderRadius: '2px'
          }} />
          {intl.formatMessage({id: 'pages.switch.ruleBuilder.title'})}
        </div>
        <div style={{
          textAlign: 'center',
          padding: '40px',
          background: 'white',
          borderRadius: '12px',
          border: '1px solid #e1e4e8'
        }}>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={createInitialGroup}
            size="large"
            style={{
              height: '40px',
              borderRadius: '8px'
            }}
          >
            {intl.formatMessage({id: 'pages.switch.ruleBuilder.createRuleGroup'})}
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div style={{
      background: '#fafbfc',
      border: '1px solid #e1e4e8',
      borderRadius: '16px',
      padding: '24px'
    }}>
      <div style={{
        fontSize: '16px',
        fontWeight: '600',
        color: '#24292f',
        marginBottom: '20px',
        display: 'flex',
        alignItems: 'center',
        gap: '8px'
      }}>
        <div style={{
          width: '4px',
          height: '16px',
          background: '#1890ff',
          borderRadius: '2px'
        }} />
        {intl.formatMessage({id: 'pages.switch.ruleBuilder.title'})}
      </div>
      <div style={{
        background: 'white',
        borderRadius: '12px',
        padding: '16px',
        border: '1px solid #e1e4e8'
      }}>
        <RuleGroup
          node={rootRule}
          onUpdate={(updatedNode) => updateRule(updatedNode)}
          onDelete={() => updateRule(undefined)}
          isRoot={true}
          factorOptions={factorOptions}
          intl={intl}
        />
      </div>
    </div>
  );
};

interface RuleGroupProps {
  node: API.RuleNode;
  onUpdate: (node: API.RuleNode) => void;
  onDelete: () => void;
  isRoot?: boolean;
  factorOptions?: API.SwitchFactorItem[];
  intl: any;
}

const RuleGroup: React.FC<RuleGroupProps> = ({ node, onUpdate, onDelete, isRoot = false, factorOptions = [], intl }) => {
  // 因子添加后ID设置为0，只有在下拉中选择具体的因子后才会有具体的ID值
  // 非协调节点nodeType为空
  const addFactor = () => {
    const newFactor: API.RuleNode = {
      id: 0,
      factor: '',
      config: ''
    };

    const updatedNode = {
      ...node,
      children: [...(node.children || []), newFactor]
    };
    onUpdate(updatedNode);
  };

  const addGroup = () => {
    const newGroup: API.RuleNode = {
      id: 0,
      nodeType: 'AND',
      children: []
    };

    const updatedNode = {
      ...node,
      children: [...(node.children || []), newGroup]
    };
    onUpdate(updatedNode);
  };

  const updateChild = (index: number, childNode: API.RuleNode) => {
    const newChildren = [...(node.children || [])];
    newChildren[index] = childNode;
    onUpdate({ ...node, children: newChildren });
  };

  const deleteChild = (index: number) => {
    const newChildren = [...(node.children || [])];
    newChildren.splice(index, 1);
    onUpdate({ ...node, children: newChildren });
  };

  const toggleOperator = () => {
    const currentType = node.nodeType;
    let newType: string;

    if (currentType === 'AND') {
      newType = 'OR';
    } else if (currentType === 'OR') {
      newType = 'AND';
    } else {
      //默认 'AND'
      newType = 'AND';
    }

    onUpdate({
      ...node,
      nodeType: newType as any
    });
  };


  return (
    // and or 都属于协调节点
    <div style={{
      border: (node.nodeType === 'AND' || node.nodeType === 'OR') ? '2px solid #1890ff' : '2px solid #fa8c16',
      background: 'white',
      borderRadius: '12px',
      padding: '20px',
      margin: '14px 0',
      boxShadow: '0 8px 32px rgba(0,0,0,0.08)',
      position: 'relative',
      transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
      minHeight: (!node.children || node.children.length === 0) ? '80px' : 'auto'
    }}>
      <div style={{
        position: 'absolute',
        top: '-16px',
        left: '16px',
        display: 'flex',
        alignItems: 'center',
        gap: '8px'
      }}>
        <div
          style={{
            background: (node.nodeType === 'AND' || node.nodeType === 'OR') ? '#1890ff' : '#fa8c16',
            color: 'white',
            padding: '6px 16px',
            borderRadius: '16px',
            fontSize: '11px',
            fontWeight: '700',
            cursor: 'pointer',
            boxShadow: '0 4px 16px rgba(0,0,0,0.15)',
            transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
            userSelect: 'none',
            letterSpacing: '0.5px'
          }}
          onClick={toggleOperator}
          onMouseEnter={(e) => {
            e.currentTarget.style.transform = 'translateY(-2px) scale(1.05)';
            e.currentTarget.style.boxShadow = '0 8px 24px rgba(0,0,0,0.2)';
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.transform = 'translateY(0) scale(1)';
            e.currentTarget.style.boxShadow = '0 4px 16px rgba(0,0,0,0.15)';
          }}
        >
          {(node.nodeType === 'AND' || node.nodeType === 'OR') ? node.nodeType : 'AND'}
        </div>

        <div
          style={{
            background: '#52c41a',
            color: 'white',
            padding: '6px 16px',
            borderRadius: '16px',
            fontSize: '11px',
            fontWeight: '700',
            cursor: 'pointer',
            boxShadow: '0 4px 16px rgba(0,0,0,0.15)',
            transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
            userSelect: 'none',
            letterSpacing: '0.5px'
          }}
          onClick={addFactor}
          onMouseEnter={(e) => {
            e.currentTarget.style.transform = 'translateY(-2px) scale(1.05)';
            e.currentTarget.style.boxShadow = '0 8px 24px rgba(0,0,0,0.2)';
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.transform = 'translateY(0) scale(1)';
            e.currentTarget.style.boxShadow = '0 4px 16px rgba(0,0,0,0.15)';
          }}
        >
          {intl.formatMessage({id: 'pages.switch.ruleBuilder.factor'})}
        </div>

        <div
          style={{
            background: '#722ed1',
            color: 'white',
            padding: '6px 16px',
            borderRadius: '16px',
            fontSize: '11px',
            fontWeight: '700',
            cursor: 'pointer',
            boxShadow: '0 4px 16px rgba(0,0,0,0.15)',
            transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
            userSelect: 'none',
            letterSpacing: '0.5px'
          }}
          onClick={addGroup}
          onMouseEnter={(e) => {
            e.currentTarget.style.transform = 'translateY(-2px) scale(1.05)';
            e.currentTarget.style.boxShadow = '0 8px 24px rgba(0,0,0,0.2)';
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.transform = 'translateY(0) scale(1)';
            e.currentTarget.style.boxShadow = '0 4px 16px rgba(0,0,0,0.15)';
          }}
        >
          {intl.formatMessage({id: 'pages.switch.ruleBuilder.ruleGroup'})}
        </div>
      </div>

      {/*定义删除按钮*/}
      {!isRoot && (
        <div style={{
          position: 'absolute',
          top: '-8px',
          right: '-8px'
        }}>
          <Button
            size="middle"
            icon={<DeleteOutlined />}
            onClick={onDelete}
            shape="circle"
            style={{
              background: '#ff4d4f',
              border: 'none',
              color: 'white',
              width: '28px',
              height: '32px',
              boxShadow: '0 4px 16px rgba(255, 107, 107, 0.4)',
              transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)'
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.transform = 'scale(1.1)';
              e.currentTarget.style.boxShadow = '0 8px 24px rgba(255, 107, 107, 0.5)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = 'scale(1)';
              e.currentTarget.style.boxShadow = '0 4px 16px rgba(255, 107, 107, 0.4)';
            }}
          />
        </div>
      )}

      {/* 子节点 */}
      {(node.children || []).map((child, index) => {
        const children = node.children || [];
        const prevChild = index > 0 ? children[index - 1] : null;
        const nextChild = index < children.length - 1 ? children[index + 1] : null;

        let marginTop = '0';
        let marginBottom = '8px';

        if (index === 0) {
          marginTop = (child.nodeType === 'AND' || child.nodeType === 'OR') ? '16px' : '12px';
        }

        if (child.nodeType === 'AND' || child.nodeType === 'OR') {
          marginBottom = '20px';
        } else if (nextChild?.nodeType === 'AND' || nextChild?.nodeType === 'OR') {
          marginBottom = '16px';
        } else {
          marginBottom = '8px';
        }

        if ((prevChild?.nodeType === 'AND' || prevChild?.nodeType === 'OR') && child.nodeType === '') {
          marginTop = '24px';
        }

        if (prevChild?.nodeType === '' && (child.nodeType === 'AND' || child.nodeType === 'OR')) {
          marginTop = '24px';
        }

        return (
          <div key={index} style={{
            marginTop,
            marginBottom
          }}>
            {(child.nodeType === 'AND' || child.nodeType === 'OR') ? (
              <RuleGroup
                node={child}
                onUpdate={(updatedChild) => updateChild(index, updatedChild)}
                onDelete={() => deleteChild(index)}
                factorOptions={factorOptions}
                intl={intl}
              />
            ) : (
              <RuleFactor
                node={child}
                onUpdate={(updatedChild) => updateChild(index, updatedChild)}
                onDelete={() => deleteChild(index)}
                factorOptions={factorOptions}
                intl={intl}
              />
            )}
          </div>
        );
      })}

      {/* 空状态提示 */}
      {(!node.children || node.children.length === 0) && (
        <div style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          height: '60px',
          color: '#8c8c8c',
          fontSize: '13px',
          fontWeight: '500',
          background: 'rgba(24, 144, 255, 0.02)',
          borderRadius: '8px',
          border: '1px dashed rgba(24, 144, 255, 0.2)',
          margin: '8px 0'
        }}>
          {intl.formatMessage({id: 'pages.switch.ruleBuilder.emptyTip'})}
        </div>
      )}

    </div>
  );
};

interface RuleFactorProps {
  node: API.RuleNode;
  onUpdate: (node: API.RuleNode) => void;
  onDelete: () => void;
  factorOptions?: API.SwitchFactorItem[];
  intl: any;
}

const RuleFactor: React.FC<RuleFactorProps> = ({ node, onUpdate, onDelete, factorOptions = [], intl }) => {
  const selectedFactor = factorOptions.find((f: API.SwitchFactorItem) => f.id === node.id);

  const updateFactor = (field: string, value: any) => {
    onUpdate({ ...node, [field]: value });
  };

  const renderValueInput = () => {
    if (!selectedFactor || !selectedFactor.factor) return null;

    return (
      <div style={{ width: 400, minWidth: 300 }}>
        <UniversalFormRenderer
          input={JSON.parse(selectedFactor.jsonSchema || '{}')}
          value={node.config}
          onChange={(value) => updateFactor('config', value)}
          readonly={false}
        />
      </div>
    );
  };

  return (
    <div style={{
      display: 'flex',
      alignItems: 'center',
      gap: '14px',
      padding: '18px',
      backgroundColor: 'white',
      borderRadius: '12px',
      border: '1px solid #e1e4e8',
      boxShadow: '0 4px 16px rgba(0,0,0,0.04)',
      margin: '10px 0',
      transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
      position: 'relative'
    }}>
      <div style={{
        width: '8px',
        height: '8px',
        borderRadius: '50%',
        background: '#1890ff',
        flexShrink: 0,
        boxShadow: '0 2px 8px rgba(24, 144, 255, 0.3)'
      }} />

      <Select
        placeholder={intl.formatMessage({id: 'pages.switch.ruleBuilder.selectFactor'})}
        value={node.id === 0 ? undefined : node.id?.toString()}
        onChange={(value) => {
          const selectedFactorId = parseInt(value);
          const selectedFactor = factorOptions.find(f => f.id === selectedFactorId);
          onUpdate({
            ...node,
            id: selectedFactorId,
            factor: selectedFactor?.factor || '',
            config: ''
          });
        }}
        style={{
          width: 160
        }}
        size="middle"
        variant="filled"
        options={factorOptions?.map((factor: API.SwitchFactorItem) => ({
          label: factor.name,
          value: factor.id.toString()
        })) || []}
      />

      {/*核心逻辑 根据json schema的配置去渲染组件，并填充json API.RuleNode的值*/}
      {renderValueInput()}

      {/*定义因子的删除按钮*/}
      <Button
        size="middle"
        icon={<DeleteOutlined />}
        onClick={onDelete}
        shape="circle"
        style={{
          flexShrink: 0,
          background: '#ff4d4f',
          border: 'none',
          color: 'white',
          width: '32px',
          height: '32px',
          boxShadow: '0 2px 8px rgba(255, 107, 107, 0.3)',
          transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)'
        }}
        onMouseEnter={(e) => {
          e.currentTarget.style.transform = 'scale(1.1)';
          e.currentTarget.style.boxShadow = '0 4px 16px rgba(255, 107, 107, 0.4)';
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.transform = 'scale(1)';
          e.currentTarget.style.boxShadow = '0 2px 8px rgba(255, 107, 107, 0.3)';
        }}
      />
    </div>
  );
};

export default RuleBuilder;
