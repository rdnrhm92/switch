import {ActionType, PageContainer, ProColumns, ProTable,} from '@ant-design/pro-components';
import {Access, FormattedMessage, useAccess, useIntl, history} from '@umijs/max';
import {Button, message, Modal, Tabs} from 'antd';
import React, {useRef, useState, useEffect} from 'react';
import JsonView from '@uiw/react-json-view';
import {PlusOutlined, EditOutlined, DiffOutlined} from "@ant-design/icons";
import {switches, pushSwitchChange} from "@/services/switch/api";
import {namespaceEnvironments} from '@/services/env/api';
import {useModel} from "@@/exports";
import DiffModal, {SwitchConfig} from './components/DiffModal';
import {useErrorHandler} from "@/utils/useErrorHandler";


const TableList: React.FC = () => {
  const [showDetail, setShowDetail] = useState<boolean>(false);
  const [showDiff, setShowDiff] = useState<boolean>(false);
  const [currentRow, setCurrentRow] = useState<API.SwitchModel>();
  const [currentSwitchConfigs, setCurrentSwitchConfigs] = useState<SwitchConfig[]>([]);
  const [environments, setEnvironments] = useState<any[]>([]);
  const actionRef = useRef<ActionType | null>(null);
  const {initialState} = useModel('@@initialState');
  const {currentUser} = initialState || {};
  const access = useAccess() as Record<string, boolean>;
  const { handleError } = useErrorHandler();

  /**
   * @en-US International configuration
   * @zh-CN 国际化配置
   * */
  const intl = useIntl();

  const [messageApi, contextHolder] = message.useMessage();

  // 加载环境列表
  const loadEnvironments = async () => {
    if (!currentUser?.select_namespace) {
      messageApi.error(intl.formatMessage({ id: 'pages.switch.message.namespaceNotFound' }));
      return;
    }

    try {
      const response = await namespaceEnvironments(currentUser.select_namespace);
      const envList = response?.data || [];
      setEnvironments(envList);
    } catch (error) {
      handleError(
        error,
        'pages.switch.message.loadEnvError',
        { showDetail: true }
      );
    }
  };

  // 检查环境是否存在
  const checkEnvironments = async () => {
    if (environments.length === 0) {
      Modal.confirm({
        title: intl.formatMessage({ id: 'pages.switch.modal.createEnvTitle' }),
        content: intl.formatMessage({ id: 'pages.switch.modal.createEnvContent' }),
        okText: intl.formatMessage({ id: 'pages.switch.modal.createEnvOk' }),
        cancelText: intl.formatMessage({ id: 'pages.switch.modal.createEnvCancel' }),
        onOk: () => {
          history.push('/config/env');
        },
      });
      return false;
    }
    return true;
  };

  // 获取环境标签颜色(带有警示作用)
  const getEnvironmentColor = (currentEnvTag: string) => {
    const envIndex = environments.findIndex(env => env.tag === currentEnvTag);
    if (envIndex === -1) return '#722ed1';

    const totalEnvs = environments.length;
    const position = envIndex + 1;

    if (position === 1) {
      return '#52c41a';
    } else if (position === totalEnvs) {
      return '#ff4d4f';
    } else if (position === totalEnvs - 1) {
      return '#fa8c16';
    } else {
      return '#1890ff';
    }
  };

  // 处理新建按钮点击
  const handleCreateClick = async () => {
    const hasEnvironments = await checkEnvironments();
    if (hasEnvironments) {
      history.push('/list/switch/create');
    }
  };


  // 获取目标环境Tag（从nextEnvInfo中获取）
  const getTargetEnvironmentTag = (nextEnvInfo: API.NextEnvInfo | undefined) => {
    if (!nextEnvInfo?.envTag) {
      return null;
    }
    // 从环境列表中找到对应的环境ID
    const targetEnv = environments.find(env => env.tag === nextEnvInfo.envTag);
    return targetEnv?.tag || null;
  };

  // 处理环境标签按钮点击
  const handleEnvTagClick = async (record: API.SwitchModel) => {
    if (!record.id) {
      messageApi.error(intl.formatMessage({ id: 'pages.switch.message.switchIdMissing' }));
      return;
    }

    const targetEnvTag = getTargetEnvironmentTag(record.nextEnvInfo);
    if (!targetEnvTag) {
      messageApi.error(intl.formatMessage({ id: 'pages.switch.message.targetEnvMissing' }));
      return;
    }

    try {
      await pushSwitchChange({
        switchId: record.id,
        targetEnvTag: targetEnvTag
      });
      messageApi.success(intl.formatMessage({ id: 'pages.switch.message.pushSuccess' }));
      // 刷新列表
      actionRef.current?.reload();
    } catch (error: any) {
      handleError(
        error,
        'pages.switch.message.pushError',
        { showDetail: true }
      );
    }
  };

  const handleDiffClick = (record: API.SwitchModel) => {
    const switchConfigs = record.switchConfigs || [];
    const newConfigs = switchConfigs
      .filter(config => config.status === 'PUBLISHED')
      .map(config => {
        return config
      });
    setCurrentSwitchConfigs(newConfigs);
    setShowDiff(true);
  };

  // 页面加载时获取环境列表用于环境开关配置的推送跟是否可以新建的判断
  useEffect(() => {
    if (currentUser?.select_namespace) {
      loadEnvironments();
    }
  }, [currentUser?.select_namespace]);

  const fetchSwitchList = async (params: any, sort: any, filter: any) => {
    try {
      const result = await switches(params, sort, filter);
      return {
        ...result,
        success: true,
      };
    } catch (error: any) {
      handleError(
        error,
        'pages.switch.message.listLoadError',
        { showDetail: true }
      );
      return {
        data: [],
        success: false,
        total: 0,
      };
    }
  };

  const columns: ProColumns<API.SwitchModel>[] = [
    {
      title: (
        <FormattedMessage
          id="pages.switch.searchTable.id"
        />
      ),
      dataIndex: 'id',
      width: 60,
      hideInSearch: true,
    },
    {
      title: (
        <FormattedMessage
          id="pages.switch.searchTable.name"
        />
      ),
      dataIndex: 'name',
      width: 120,
      render: (dom, entity) => {
        return (
          <a
            onClick={() => {
              setCurrentRow(entity);
              setShowDetail(true);
            }}
          >
            {dom}
          </a>
        );
      },
    },
    {
      title: (
        <FormattedMessage
          id="pages.switch.searchTable.description"
        />
      ),
      dataIndex: 'description',
      width: 150,
      ellipsis: true,
    },
    {
      title: (
        <FormattedMessage
          id="pages.switch.searchTable.version"
        />
      ),
      dataIndex: 'version',
      width: 80,
      hideInSearch: true,
    },
    {
      title: (
        <FormattedMessage
          id="pages.switch.searchTable.useCache"
        />
      ),
      dataIndex: 'useCache',
      width: 80,
      hideInSearch: true,
      render: (_, record) => {
        return record.useCache
          ? intl.formatMessage({ id: 'pages.switch.searchTable.yes' })
          : intl.formatMessage({ id: 'pages.switch.searchTable.no' });
      },
    },
    {
      title: (
        <FormattedMessage
          id="pages.switch.searchTable.createdBy"
        />
      ),
      dataIndex: 'createdBy',
      width: 80,
      hideInSearch: true,
    },
    {
      title: (
        <FormattedMessage
          id="pages.switch.searchTable.createTime"
        />
      ),
      dataIndex: 'createTime',
      width: 140,
      hideInSearch: true,
      render: (_, record) => {
        if (!record.createTime) return '-';
        return record.createTime.replace('T', ' ').substring(0, 19);
      },
    },
    {
      title: (
        <FormattedMessage
          id="pages.switch.searchTable.updateBy"
        />
      ),
      dataIndex: 'updateBy',
      width: 80,
      hideInSearch: true,
    },
    {
      title: (
        <FormattedMessage
          id="pages.switch.searchTable.updateTime"
        />
      ),
      dataIndex: 'updateTime',
      width: 140,
      hideInSearch: true,
      render: (_, record) => {
        if (!record.updateTime) return '-';
        return record.updateTime.replace('T', ' ').substring(0, 19);
      },
    },
    {
      title: (
        <FormattedMessage
          id="pages.switch.searchTable.option"
        />
      ),
      dataIndex: 'option',
      valueType: 'option',
      width: 200,
      fixed: 'right',
      render: (_, record) => [
        (() => {
          const { nextEnvInfo } = record;

          if (!nextEnvInfo) {
            // 没有下一个环境（最后一个环境），显示不可点击按钮，保持原来的颜色
            return (
              <Button
                key="env-tag"
                size="small"
                disabled={true}
                style={{
                  backgroundColor: getEnvironmentColor(record.currentEnvTag || ''),
                  color: '#fff',
                  border: 'none',
                  marginRight: 8,
                  fontSize: '12px',
                  height: '24px',
                  lineHeight: '26px',
                  borderRadius: '12px',
                  minWidth: '60px',
                  padding: '0 12px',
                  cursor: 'not-allowed',
                  opacity: 0.6
                }}
              >
                {record.currentEnvTag}
              </Button>
            );
          }

          // 没有权限或按钮本身被禁用
          const isDisabled = !access['switches:push'] || nextEnvInfo.buttonDisabled;

          return (
            <Button
              key="env-tag"
              size="small"
              disabled={isDisabled}
              onClick={() => handleEnvTagClick(record)}
              style={{
                backgroundColor: getEnvironmentColor(nextEnvInfo.envTag),
                color: '#fff',
                border: 'none',
                marginRight: 8,
                fontSize: '12px',
                height: '24px',
                lineHeight: '26px',
                borderRadius: '12px',
                minWidth: '60px',
                padding: '0 12px',
                cursor: isDisabled ? 'not-allowed' : 'pointer',
                opacity: isDisabled ? 0.6 : 1
              }}
            >
              {nextEnvInfo.buttonText}
            </Button>
          );
        })(),
        <Button
          key="diff"
          type="link"
          size="small"
          icon={<DiffOutlined />}
          onClick={() => handleDiffClick(record)}
        >
          {intl.formatMessage({ id: 'pages.switch.searchTable.diff' })}
        </Button>,
        <Access accessible={!!access['switches:edit']} key="edit">
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => {
              history.push(`/list/switch/edit/${record.id}`);
            }}
          >
            {intl.formatMessage({ id: 'pages.switch.searchTable.edit' })}
          </Button>
        </Access>
      ],
    },
  ];

  const tabItems = [
    {
      key: 'switch-list',
      label: intl.formatMessage({ id: 'pages.switch.searchTable.tabList' }),
      children: (
        <>
          <ProTable<API.SwitchModel, API.PageParams>
            headerTitle={intl.formatMessage({
              id: 'pages.switch.searchTable.title',
            })}
            actionRef={actionRef}
            rowKey="id"
            search={{
              labelWidth: 'auto',
              span: {
                xs: 24,
                sm: 12,
                md: 8,
                lg: 8,
                xl: 6,
                xxl: 6,
              },
            }}
            toolBarRender={() => [
              <Access accessible={!!access['switches:create']} key="create">
                <Button
                  type={'primary'}
                  icon={<PlusOutlined/>}
                  onClick={handleCreateClick}
                >
                  <FormattedMessage
                    id="pages.switch.searchTable.createIcon"/>
                </Button>
              </Access>
            ]}
            request={fetchSwitchList}
            columns={columns}
            scroll={{ x: 1200 }}
            size="small"
          />

          {showDetail && (
            <Modal
              title={intl.formatMessage({ id: 'pages.switch.searchTable.detail' })}
              open={showDetail}
              onCancel={() => setShowDetail(false)}
              footer={null}
              width={800}
              centered
              destroyOnHidden={true}
            >
              <JsonView
                value={currentRow}
                style={{
                  backgroundColor: '#f6f8fa',
                  padding: '16px',
                  borderRadius: '6px'
                }}
                collapsed={false}
              />
            </Modal>
          )}

          <DiffModal
            open={showDiff}
            onCancel={() => setShowDiff(false)}
            switchConfigs={currentSwitchConfigs}
            leftTitle={intl.formatMessage({ id: 'pages.switch.searchTable.leftEnv' })}
            rightTitle={intl.formatMessage({ id: 'pages.switch.searchTable.rightEnv' })}
          />
        </>
      )
    },
  ];

  return (
    <PageContainer>
      {contextHolder}
      <Tabs items={tabItems} />
    </PageContainer>
  );
};

export default TableList;

