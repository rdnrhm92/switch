import {Form, Card, Space, Button, Select, Input, List, Tag, message} from 'antd';
import React, {useState, useEffect} from 'react';
import {PlusOutlined, MinusCircleOutlined, SearchOutlined, UserOutlined, ExclamationCircleOutlined, CheckCircleOutlined} from '@ant-design/icons';
import {namespaceEnvironments} from '@/services/env/api';
import {userPermissionsList} from '@/services/switch/api';
import {useModel} from '@@/exports';
import { parseApproverUsers } from '@/utils/approvalUtils';
import {useErrorHandler} from "@/utils/useErrorHandler";
import {useIntl} from "@umijs/max";

export type ApproverConfigProps = {
  disabled?: boolean; // 控制整个组件的可用/禁用状态
};

const ApproverConfig: React.FC<ApproverConfigProps> = ({ disabled = false }) => {
  const intl = useIntl();
  const [messageApi, contextHolder] = message.useMessage();
  const [envOptions, setEnvOptions] = useState<{label: string, value: string}[]>([]);
  const [searchTexts, setSearchTexts] = useState<{[key: number]: string}>({});
  const [availableApprovers, setAvailableApprovers] = useState<API.UserPermissionsListItem[]>([]);
  const [loadingUsers, setLoadingUsers] = useState<boolean>(false);
  const {initialState} = useModel('@@initialState');
  const {currentUser} = initialState || {};
  const { handleError } = useErrorHandler();

  // 获取环境列表
  useEffect(() => {
    const fetchEnvironments = async () => {
      if ((currentUser as any)?.select_namespace) {
        try {
          const response = await namespaceEnvironments((currentUser as any).select_namespace);

          if (response?.data) {
            const options = response.data.map((env: API.EnvItem) => ({
              label: env.name,
              value: env.tag
            }));
            setEnvOptions(options);
          }
        } catch (error) {
          handleError(
            error,
            'pages.env.list.fail',
            { showDetail: true }
          );
        }
      }
    };

    if (currentUser) {
      fetchEnvironments();
    }
  }, [currentUser, (currentUser as any)?.select_namespace]);

  // 获取用户列表
  useEffect(() => {
    const fetchUsers = async () => {
      setLoadingUsers(true);
      try {
        const response = await userPermissionsList({
          username: ''
        });
        if (Array.isArray(response)) {
          setAvailableApprovers(response);
        } else if (response) {
          setAvailableApprovers([]);
        }
      } catch (error) {
        handleError(
          error,
          'pages.env.userList.fail',
          { showDetail: true }
        );
      } finally {
        setLoadingUsers(false);
      }
    };

    fetchUsers();
  }, []);

  // 获取特定环境的过滤审批人
  const getFilteredApprovers = (fieldKey: number) => {
    const searchText = searchTexts[fieldKey] || '';
    return availableApprovers.filter(approver =>
      approver.userInfo.username.toLowerCase().includes(searchText.toLowerCase())
    );
  };

  // 更新特定环境的搜索文本
  const updateSearchText = (fieldKey: number, value: string) => {
    setSearchTexts(prev => ({
      ...prev,
      [fieldKey]: value
    }));
  };

  // 检查用户是否有审批权限
  const hasApprovalPermission = (approver: API.UserPermissionsListItem) => {
    return approver.userPermissions && approver.userPermissions.some(permission =>
      permission.name === 'approvals:approve'
    );
  };

  // 添加审批人到表单
  const addApprover = (form: any, fieldName: number, fieldKey: number, approverId: number) => {
    if (disabled) return; // 禁用状态下不允许操作

    const approver = availableApprovers.find(a => a.userInfo.id === approverId);
    if (!approver) return;

    if (!hasApprovalPermission(approver)) {
      messageApi.warning(intl.formatMessage(
        {id: 'pages.switch.approver.noPermissionWarning'},
        {username: approver.userInfo.username}
      ));
      return;
    }

    const currentApprovers = getApproverUsers(form, fieldName);
    if (!currentApprovers.includes(approverId)) {
      form.setFieldValue(['createSwitchApproversReq', fieldName, 'approverUsers'], [...currentApprovers, approverId]);
    }
    updateSearchText(fieldKey, '');
  };

  const getApproverUsers = (form: any, fieldName: number) => {
    const currentApprovers = form.getFieldValue(['createSwitchApproversReq', fieldName, 'approverUsers']);
    return parseApproverUsers(currentApprovers);
  };

  // 移除审批人
  const removeApprover = (form: any, fieldName: number, approverId: number) => {
    const currentApprovers = getApproverUsers(form, fieldName);
    form.setFieldValue(['createSwitchApproversReq', fieldName, 'approverUsers'],
      currentApprovers.filter((id: number) => id !== approverId)
    );
  };

  // 清理搜索文本状态
  const cleanupSearchText = (fieldKey: number) => {
    setSearchTexts(prev => {
      const newState = { ...prev };
      delete newState[fieldKey];
      return newState;
    });
  };

  // 获取可用的环境选项（过滤掉已选择的环境）
  const getAvailableEnvOptions = (form: any, currentFieldName: number) => {
    const allApprovers = form.getFieldValue('createSwitchApproversReq') || [];
    const selectedEnvTags = allApprovers
      .map((item: any, index: number) => index !== currentFieldName ? item?.envTag : null)
      .filter((tag: string | null) => tag != null);

    return envOptions.filter(option => !selectedEnvTags.includes(option.value));
  };

  return (
    <>
      {contextHolder}
      <Card
        title={intl.formatMessage({id: 'pages.switch.approver.title'})}
        size="small"
        style={{ marginTop: '20px', marginBottom: '60px' }}
      >
        <Form.List name="createSwitchApproversReq" key={`form-list-${envOptions.length}`}>
        {(fields, { add, remove }) => (
          <>
            {fields.map(({ key, name, ...restField }) => (
              <Space key={key} style={{ display: 'block', marginBottom: 24, padding: 16, border: '1px solid #f0f0f0', borderRadius: 6 }} align="baseline">
                <div style={{ display: 'flex', alignItems: 'center', marginBottom: 16 }}>
                  <Form.Item noStyle shouldUpdate={(prevValues, currentValues) => {
                    return prevValues.createSwitchApproversReq !== currentValues.createSwitchApproversReq;
                  }}>
                    {(form) => (
                      <Form.Item
                        {...restField}
                        name={[name, 'envTag']}
                        rules={[{ required: true, message: intl.formatMessage({id: 'pages.switch.approver.selectEnvRequired'}) }]}
                        style={{ marginBottom: 0, marginRight: 16 }}
                      >
                        <Select
                          placeholder={intl.formatMessage({id: 'pages.switch.approver.selectEnv'})}
                          style={{ width: 200 }}
                          options={getAvailableEnvOptions(form, name)}
                          showSearch
                          allowClear
                          disabled={disabled}
                          filterOption={(input, option) =>
                            (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
                          }
                          notFoundContent={envOptions.length === 0 ? intl.formatMessage({id: 'pages.switch.approver.loadingEnv'}) : intl.formatMessage({id: 'pages.switch.approver.noEnvData'})}
                        />
                      </Form.Item>
                    )}
                  </Form.Item>
                  <Button
                    type="text"
                    danger
                    icon={<MinusCircleOutlined />}
                    disabled={disabled}
                    onClick={() => {
                      cleanupSearchText(key);
                      remove(name);
                    }}
                  >
                    {intl.formatMessage({id: 'pages.switch.approver.deleteApproval'})}
                  </Button>
                </div>

                <div>
                  <div style={{ marginBottom: 12 }}>
                    <Input
                      placeholder={intl.formatMessage({id: 'pages.switch.approver.searchApprover'})}
                      prefix={<SearchOutlined />}
                      value={searchTexts[key] || ''}
                      onChange={(e) => updateSearchText(key, e.target.value)}
                      disabled={disabled}
                      style={{ width: '100%' }}
                    />
                  </div>

                  <Form.Item noStyle shouldUpdate>
                    {(form) => {
                      const currentSearchText = searchTexts[key] || '';
                      const filteredApprovers = getFilteredApprovers(key);

                      return (
                        <>
                          {currentSearchText && (
                            <div style={{ marginBottom: 16, maxHeight: 200, overflowY: 'auto', border: '1px solid #f0f0f0', borderRadius: 4 }}>
                              <List
                                size="small"
                                dataSource={filteredApprovers}
                                renderItem={(approver: API.UserPermissionsListItem) => {
                                  const canApprove = hasApprovalPermission(approver);
                                  return (
                                    <List.Item
                                      style={{
                                        cursor: (canApprove && !disabled) ? 'pointer' : 'not-allowed',
                                        padding: '8px 12px',
                                        opacity: (canApprove && !disabled) ? 1 : 0.6,
                                        backgroundColor: (canApprove && !disabled) ? 'transparent' : '#f5f5f5'
                                      }}
                                      onClick={() => !disabled && addApprover(form, name, key, approver.userInfo.id)}
                                      onMouseEnter={(e) => {
                                        if (canApprove && !disabled) {
                                          e.currentTarget.style.backgroundColor = '#f5f5f5';
                                        }
                                      }}
                                      onMouseLeave={(e) => {
                                        if (canApprove && !disabled) {
                                          e.currentTarget.style.backgroundColor = 'transparent';
                                        } else {
                                          e.currentTarget.style.backgroundColor = '#f5f5f5';
                                        }
                                      }}
                                    >
                                      <List.Item.Meta
                                        avatar={
                                          canApprove ?
                                            <CheckCircleOutlined style={{ color: '#52c41a' }} /> :
                                            <ExclamationCircleOutlined style={{ color: '#faad14' }} />
                                        }
                                        title={
                                          <span style={{ color: canApprove ? '#000' : '#999' }}>
                                            {approver.userInfo.username}
                                            {!canApprove && <span style={{ color: '#ff4d4f', marginLeft: 8 }}>({intl.formatMessage({id: 'pages.switch.approver.noApprovalPermission'})})</span>}
                                          </span>
                                        }
                                        description={
                                          <span style={{ color: canApprove ? '#666' : '#999' }}>
                                            {intl.formatMessage({id: 'pages.switch.approver.userId'})}: {approver.userInfo.id} | {intl.formatMessage({id: 'pages.switch.approver.permission'})}: { canApprove ? intl.formatMessage({id: 'pages.switch.approver.canApprove'}) : intl.formatMessage({id: 'pages.switch.approver.noPermission'})}
                                            {!canApprove && <div style={{ color: '#ff4d4f', marginTop: 4 }}>{intl.formatMessage({id: 'pages.switch.approver.missingPermission'})}</div>}
                                          </span>
                                        }
                                      />
                                    </List.Item>
                                  );
                                }}
                              />
                            </div>
                          )}
                        </>
                      );
                    }}
                  </Form.Item>

                  <Form.Item
                    {...restField}
                    name={[name, 'approverUsers']}
                    rules={[{ required: true, message: intl.formatMessage({id: 'pages.switch.approver.selectApproverRequired'}) }]}
                  >
                    <Form.Item noStyle shouldUpdate>
                      {(form) => {
                        const selectedApprovers = form.getFieldValue(['createSwitchApproversReq', name, 'approverUsers']) || [];
                        const selectedApproverData = availableApprovers.filter(approver =>
                          selectedApprovers.includes(approver.userInfo.id)
                        );

                        return (
                          <div>
                            <div style={{ marginBottom: 8, fontWeight: 500 }}>{intl.formatMessage({id: 'pages.switch.approver.selectedApprovers'})}</div>
                            {selectedApproverData.length === 0 ? (
                              <div style={{ color: '#999', fontStyle: 'italic' }}>{intl.formatMessage({id: 'pages.switch.approver.noApproverSelected'})}</div>
                            ) : (
                              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
                                {selectedApproverData.map(approver => (
                                  <Tag
                                    key={approver.userInfo.id}
                                    closable={!disabled}
                                    onClose={() => removeApprover(form, name, approver.userInfo.id)}
                                    style={{ marginBottom: 4 }}
                                  >
                                    <UserOutlined style={{ marginRight: 4 }} />
                                    {approver.userInfo.username} ({approver.userPermissions && approver.userPermissions.length > 0 ? intl.formatMessage({id: 'pages.switch.approver.canApprove'}) : intl.formatMessage({id: 'pages.switch.approver.noPermission'})})
                                  </Tag>
                                ))}
                              </div>
                            )}
                          </div>
                        );
                      }}
                    </Form.Item>
                  </Form.Item>
                </div>
              </Space>
            ))}
            <Form.Item noStyle shouldUpdate={(prevValues, currentValues) => {
              return prevValues.createSwitchApproversReq !== currentValues.createSwitchApproversReq;
            }}>
              {(form) => {
                const allApprovers = form.getFieldValue('createSwitchApproversReq') || [];
                const selectedEnvTags = allApprovers
                  .map((item: any) => item?.envTag)
                  .filter((tag: string | null) => tag != null);
                const hasAvailableEnv = selectedEnvTags.length < envOptions.length;

                return (
                  <Form.Item>
                    <Button
                      type="dashed"
                      onClick={() => add()}
                      block
                      icon={<PlusOutlined />}
                      disabled={disabled || !hasAvailableEnv}
                    >
                      {intl.formatMessage({id: 'pages.switch.approver.addEnv'})}
                    </Button>
                  </Form.Item>
                );
              }}
            </Form.Item>
          </>
        )}
      </Form.List>
    </Card>
    </>
  );
};

export default ApproverConfig;
