import React, { useState, useEffect } from 'react';
import { Modal, Row, Col, Select } from 'antd';
import JsonView from '@uiw/react-json-view';
import {useIntl} from "@umijs/max";

export interface SwitchConfig {
  envTag: string;
  [key: string]: any;
}

export interface DiffModalProps {
  open: boolean;
  onCancel: () => void;
  switchConfigs: SwitchConfig[];
  leftTitle?: string;
  rightTitle?: string;
}

const DiffModal: React.FC<DiffModalProps> = ({
  open,
  onCancel,
  switchConfigs,
  leftTitle,
  rightTitle
}) => {
  const intl = useIntl();
  const [leftEnvTag, setLeftEnvTag] = useState<string>('');
  const [rightEnvTag, setRightEnvTag] = useState<string>('');
  const [envOptions, setEnvOptions] = useState<{label: string, value: string}[]>([]);

  useEffect(() => {
    if (switchConfigs && switchConfigs.length > 0) {
      const options = switchConfigs.map(config => ({
        label: config.envTag,
        value: config.envTag
      }));
      setEnvOptions(options);

      if (options.length > 0) {
        setLeftEnvTag(options[0].value);
        setRightEnvTag(options.length > 1 ? options[1].value : options[0].value);
      }
    } else {
      setEnvOptions([]);
      setLeftEnvTag('');
      setRightEnvTag('');
    }
  }, [switchConfigs]);

  const getConfigByEnvTag = (envTag: string) => {
    if (!switchConfigs || !envTag) return {};
    const config = switchConfigs.find(c => c.envTag === envTag);
    return config || {};
  };

  const leftData = getConfigByEnvTag(leftEnvTag);
  const rightData = getConfigByEnvTag(rightEnvTag);

  return (
    <Modal
      title={intl.formatMessage({id: 'pages.switch.diff.title'})}
      open={open}
      onCancel={onCancel}
      footer={null}
      width={1400}
      centered
      destroyOnHidden={true}
      style={{ top: 20 }}
    >
      <Row gutter={16} style={{ height: '75vh' }}>
        <Col span={12} style={{ height: '100%' }}>
          <div style={{
            height: '100%',
            border: '1px solid #d9d9d9',
            borderRadius: '6px',
            display: 'flex',
            flexDirection: 'column'
          }}>
            <div style={{
              padding: '12px 16px',
              backgroundColor: '#fafafa',
              borderBottom: '1px solid #d9d9d9',
              fontWeight: 'bold',
              fontSize: '14px',
              flexShrink: 0
            }}>
              {leftTitle || intl.formatMessage({id: 'pages.switch.diff.leftConfig'})}
            </div>
            <div style={{
              flex: 1,
              overflow: 'auto',
              padding: '16px',
              minHeight: 0
            }}>
              <JsonView
                value={leftData}
                style={{
                  backgroundColor: 'transparent',
                  fontSize: '13px'
                }}
                collapsed={false}
                displayDataTypes={false}
                enableClipboard={false}
              />
            </div>
            <div style={{
              padding: '12px 16px',
              borderTop: '1px solid #d9d9d9',
              backgroundColor: '#fafafa',
              flexShrink: 0
            }}>
              <Select
                placeholder={intl.formatMessage({id: 'pages.switch.diff.selectEnv'})}
                value={leftEnvTag}
                onChange={setLeftEnvTag}
                options={envOptions}
                style={{ width: '100%' }}
                showSearch
                filterOption={(input, option) =>
                  (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
                }
              />
            </div>
          </div>
        </Col>

        <Col span={12} style={{ height: '100%' }}>
          <div style={{
            height: '100%',
            border: '1px solid #d9d9d9',
            borderRadius: '6px',
            display: 'flex',
            flexDirection: 'column'
          }}>
            <div style={{
              padding: '12px 16px',
              backgroundColor: '#fafafa',
              borderBottom: '1px solid #d9d9d9',
              fontWeight: 'bold',
              fontSize: '14px',
              flexShrink: 0
            }}>
              {rightTitle || intl.formatMessage({id: 'pages.switch.diff.rightConfig'})}
            </div>
            <div style={{
              flex: 1,
              overflow: 'auto',
              padding: '16px',
              minHeight: 0
            }}>
              <JsonView
                value={rightData}
                style={{
                  backgroundColor: 'transparent',
                  fontSize: '13px'
                }}
                collapsed={false}
                displayDataTypes={false}
                enableClipboard={false}
              />
            </div>
            <div style={{
              padding: '12px 16px',
              borderTop: '1px solid #d9d9d9',
              backgroundColor: '#fafafa',
              flexShrink: 0
            }}>
              <Select
                placeholder={intl.formatMessage({id: 'pages.switch.diff.selectEnv'})}
                value={rightEnvTag}
                onChange={setRightEnvTag}
                options={envOptions}
                style={{ width: '100%' }}
                showSearch
                filterOption={(input, option) =>
                  (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
                }
              />
            </div>
          </div>
        </Col>
      </Row>
    </Modal>
  );
};

export default DiffModal;
