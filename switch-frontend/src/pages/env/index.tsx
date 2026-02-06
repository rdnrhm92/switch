import {ActionType, PageContainer, ProColumns, ProTable,} from '@ant-design/pro-components';
import {Access, FormattedMessage, useAccess, useIntl} from '@umijs/max';
import {Button, message, Modal} from 'antd';
import React, {useRef, useState} from 'react';
import {envs, publishEnv} from "@/services/env/api";
import JsonView from '@uiw/react-json-view';
import CreateUpdateForm from "./components/CreateUpdateForm";
import {PlusOutlined} from "@ant-design/icons";
import { useErrorHandler } from '@/utils/useErrorHandler';


const TableList: React.FC = () => {
  const [showDetail, setShowDetail] = useState<boolean>(false);
  const [currentRow, setCurrentRow] = useState<API.EnvItem>();
  const actionRef = useRef<ActionType | null>(null);
  const [jumpData] = useState<API.NamespaceItem>();
  const access = useAccess() as Record<string, boolean>;
  const {handleError} = useErrorHandler();


  /**
   * @en-US International configuration
   * @zh-CN 国际化配置
   * */
  const intl = useIntl();

  const [messageApi, contextHolder] = message.useMessage();

  const publish = async (item: API.EnvItem) => {
    const hide = messageApi.loading(intl.formatMessage({ id: 'pages.env.action.publishing' }), 0);
    try {
      const result = await publishEnv({id: item.id});
      hide();
      if (result && result.code == 0) {
        messageApi.success(intl.formatMessage({ id: 'pages.env.action.publishSuccess'}));
        actionRef.current?.reload();
      }
    } catch (error: any) {
      hide();
      handleError(error, 'pages.env.action.publishError', {
        showDetail: true
      });
      return {
        data: [],
        success: false,
        total: 0,
      };
    }
  };


  const fetchEnvList = async (params: any, sort: any, filter: any) => {
    try {
      const result = await envs(params, sort, filter);

      // 驱动类型 前端解析
      if (result.data && Array.isArray(result.data)) {
        result.data = result.data.map((env: any) => {
          if (env.drivers && Array.isArray(env.drivers)) {
            env.drivers = env.drivers.map((driver: any) => ({
              ...driver,
              driverType: driver.driverType ? driver.driverType.split('_')[0] : driver.driverType
            }));
          }
          return env;
        });
      }

      return {
        ...result,
        success: true,
      };
    } catch (error: any) {
      handleError(error, 'pages.env.message.listLoadError', {
        showDetail: true
      });
      return {
        data: [],
        success: false,
        total: 0,
      };
    }
  };

  const columns: ProColumns<API.EnvItem>[] = [
    {
      title: (
        <FormattedMessage
          id="pages.namespace.searchTable.id"
        />
      ),
      dataIndex: 'id',
      width: 80,
      hideInSearch: true,
    },
    {
      title: (
        <FormattedMessage
          id="pages.env.searchTable.name"
        />
      ),
      dataIndex: 'name',
      width: 150,
      fieldProps: {
        style: {
          width: '150px',
        },
      },
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
          id="pages.env.searchTable.tag"
        />
      ),
      dataIndex: 'tag',
      width: 150,
      valueType: 'textarea',
      fieldProps: {
        style: {
          width: '150px',
        },
      },
    },
    {
      title: (
        <FormattedMessage
          id="pages.env.searchTable.publishOrder"
        />
      ),
      dataIndex: 'publish_order',
      width: 150,
      hideInSearch: true,
      valueType: 'digit',
      fieldProps: {
        style: {
          width: '150px',
        },
      },
    },
    {
      title: (
        <FormattedMessage
          id="pages.env.searchTable.description"
        />
      ),
      dataIndex: 'description',
      hideInForm: true,
      width: 350,
      fieldProps: {
        style: {
          width: '150px',
        },
      },
    },
    {
      title: (
        <FormattedMessage
          id="pages.env.searchTable.publish"
        />
      ),
      dataIndex: 'publish',
      hideInForm: true,
      hideInSearch: true,
      width: 350,
      fieldProps: {
        style: {
          width: '150px',
        },
      },
      render: (_, record) => {
        return record.publish
          ? intl.formatMessage({id: 'pages.env.status.published'})
          : intl.formatMessage({id: 'pages.env.status.unpublished'});
      },
    },
    {
      title: (
        <FormattedMessage
          id="pages.namespace.searchTable.createBy"
        />
      ),
      dataIndex: 'createdBy',
      sorter: true,
      hideInForm: true,
      width: 150,
      fieldProps: {
        style: {
          width: '150px',
        },
      },
    },
    {
      title: (
        <FormattedMessage
          id="pages.namespace.searchTable.createTime"
        />
      ),
      sorter: true,
      hideInForm: true,
      width: 170,
      dataIndex: 'createTime',
      key: 'updatedAt',
      valueType: 'dateTimeRange',
      render: (_, record) => {
        if (!record.createTime) return '-';
        return record.createTime.replace('T', ' ').substring(0, 19);
      },
      fieldProps: {
        style: {
          width: '360px',
        },
      },
    },
    {
      title: (
        <FormattedMessage
          id="pages.namespace.searchTable.updateBy"
        />
      ),
      dataIndex: 'updateBy',
      sorter: true,
      hideInForm: true,
      hideInSearch: true,
      width: 150,
      fieldProps: {
        style: {
          width: '150px',
        },
      },
    },
    {
      title: (
        <FormattedMessage
          id="pages.namespace.searchTable.updateTime"
        />
      ),
      sorter: true,
      hideInForm: true,
      hideInSearch: true,
      width: 170,
      dataIndex: 'updateTime',
      render: (_, record) => {
        if (!record.createTime) return '-';
        return record.createTime.replace('T', ' ').substring(0, 19);
      },
      fieldProps: {
        style: {
          width: '360px',
        },
      },
    },
    ...((access['envs:edit'] || access['envs:publish']) ? [
      {
        title: (
          <FormattedMessage
            id="pages.system.titleOption"
          />
        ),
        dataIndex: 'option',
        valueType: 'option' as const,
        width: 170,
        render: (_:any, record:API.EnvItem) => [
          <Access key="edit" accessible={!!access['envs:edit']}>
            <CreateUpdateForm
              isCreate={false}
              trigger={
                <a>
                  <FormattedMessage
                    id="pages.env.searchTable.edit"
                  />
                </a>
              }
              key="edit"
              reload={() => actionRef.current?.reload()}
              values={record}
            />
          </Access>,
          <Access key="publish" accessible={!!access['envs:publish']}>
            {!record.publish && (
              <a
                key="publish"
                onClick={() => {
                  publish(record)
                }}
              >
                <FormattedMessage
                  id="pages.env.searchTable.publishIcon"
                />
              </a>
            )}
          </Access>
        ],
      },
    ]: []),
  ];

  return (
    <PageContainer>
      {contextHolder}
      <ProTable<API.EnvItem, API.PageParams>
        headerTitle={intl.formatMessage({
          id: 'pages.env.searchTable.title',
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
          <Access accessible={!!access['envs:create']}>
            <>
              <CreateUpdateForm
                isCreate={true}
                selectNamespace={jumpData?.tag}
                trigger={
                  <Button type={'primary'} icon={<PlusOutlined/>}>
                    <FormattedMessage
                      id="pages.env.searchTable.create"/>
                  </Button>
                }
                key="create"
                reload={() => actionRef.current?.reload()}
              />
            </>
          </Access>
        ]}
        request={fetchEnvList}
        columns={columns}
      />

      {showDetail && (
        <Modal
          title={intl.formatMessage({id: 'pages.env.detail.title'})}
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

    </PageContainer>
  );
};

export default TableList;

