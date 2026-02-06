import {SelectLang, useIntl, useModel} from '@umijs/max';
import {Alert, App, Button, Card, Modal, Select, Spin, Typography} from 'antd';
import {createStyles} from 'antd-style';
import React, {useState} from 'react';
import {Footer} from '@/components';
import {FormattedMessage} from "@@/plugin-locale";
import CreateUpdateForm from "@/pages/namespace/components/CreateUpdateForm";
import {applyToJoin, createNamespace} from '@/services/namespace/api';
import {refreshToken} from "@/pages/user/useUser";
import {indexPath} from "@/path";
import {useErrorHandler} from '@/utils/useErrorHandler';

const useStyles = createStyles(({token}) => {
  return {
    action: {
      marginLeft: '8px',
      color: 'rgba(0, 0, 0, 0.2)',
      fontSize: '24px',
      verticalAlign: 'middle',
      cursor: 'pointer',
      transition: 'color 0.3s',
      '&:hover': {
        color: token.colorPrimaryActive,
      },
    },
    lang: {
      width: 42,
      height: 42,
      lineHeight: '42px',
      position: 'fixed',
      right: 16,
      borderRadius: token.borderRadius,
      ':hover': {
        backgroundColor: token.colorBgTextHover,
      },
    },
    container: {
      display: 'flex',
      flexDirection: 'column',
      height: '100vh',
      overflow: 'auto',
      backgroundImage: "url('/bk.png')",
      backgroundSize: '100% 100%',
    },
  };
});

const Lang = () => {
  const {styles} = useStyles();

  return (
    <div className={styles.lang} data-lang>
      {SelectLang && <SelectLang/>}
    </div>
  );
};

const LoginMessage: React.FC<{
  content: string;
}> = ({content}) => {
  return (
    <Alert
      style={{
        marginBottom: 24,
      }}
      message={content}
      type="error"
      showIcon
    />
  );
};

const Login: React.FC = () => {
  const {initialState, setInitialState} = useModel('@@initialState');
  const {styles} = useStyles();
  const {message} = App.useApp();
  const intl = useIntl();
  const { handleError } = useErrorHandler();


  const [loading, setLoading] = useState(false);
  const [allNamespaces, setAllNamespaces] = useState<Select[]>([]);
  const [viewMode, setViewMode] = useState<'select' | 'create' | 'join'>('select');
  const [selectedNamespaceId, setSelectedNamespaceId] = useState<number | undefined>();
  const {currentUser} = initialState || {};
  console.log('打印一下数据：', currentUser)

  const handleCreateNamespace = async (params: Partial<API.NamespaceCreateUpdate>) => {
    setLoading(true);
    try {
      let createRes = await createNamespace(params)
      console.log('创建结果：',createRes)
      if (createRes && createRes.code == 0 && createRes.data) {
        const token = await initialState?.refreshToken?.({userId: currentUser?.id, selectNamespace: createRes.data.tag});
        if (token && token.token) {
          //这里更新token
          refreshToken(token.token)
          window.location.href = indexPath;
          return;
        } else {
          message.error(intl.formatMessage({ id: 'pages.index.namespace.createFailed' }));
        }
      }
    } catch (error) {
      handleError(
        error,
        'pages.index.namespace.createFailed',
        { showDetail: true }
      );
    } finally {
      setLoading(false);
    }
  };

  // 处理选择命名空间的逻辑
  const handleSelectNamespace = async (select: Select) => {
    setLoading(true);
    try {
      console.log('选择工作空间ID:', select.id);
      //刷新token
      const token = await initialState?.refreshToken?.({userId: currentUser?.id, selectNamespace: select.tag});
      if (token && token.token) {
        //这里更新token
        refreshToken(token.token)
        window.location.href = indexPath;
      }else{
        message.error(intl.formatMessage({ id: 'pages.index.namespace.selectFailed' }));
      }
    } catch (error) {
      handleError(
        error,
        'pages.index.namespace.selectFailed',
        { showDetail: true }
      );
    } finally {
      setLoading(false);
    }
  };

  // 处理申请加入命名空间的逻辑
  const handleRequestToJoin = async (select: Select) => {
    message.loading({content: intl.formatMessage({ id: 'pages.index.namespace.applyingToJoin' }, { name: select.name, tag: select.tag }), key: 'joinRequest'});
    try {
      let res = await applyToJoin({namespaceTag: select.tag});
      console.log(res)
      if (res && res.code == 0){
        fetchAllNamespaces();
        message.success({content: intl.formatMessage({ id: 'pages.index.namespace.applySuccess' }, { name: select.name, tag: select.tag }), key: 'joinRequest'});
      }else{
        message.error({content: intl.formatMessage({ id: 'pages.index.namespace.applyFailed' }), key: 'joinRequest'});
      }
    } catch (error) {
      handleError(
        error,
        'pages.index.namespace.applyFailed',
        { showDetail: true, key: "joinRequest" }
      );
    }
  };

  const fetchAllNamespaces = async () => {
    console.log('准备触发')
    setLoading(true);
    try {
      const namespaces = await initialState?.selectNamespace?.({search: "", all: true});
      if (namespaces) {
        const allOptions = namespaces.map(item => ({
          status: item.status,
          id: item.id,
          name: `${item.name}`,
          tag: `${item.tag}`
        }));
        setAllNamespaces(allOptions);
        console.log('获取到的空间列表', allNamespaces)
      }
    } catch (error) {
      handleError(
        error,
        'pages.index.namespace.fetchListFailed',
        { showDetail: true }
      );
    } finally {
      setLoading(false);
    }
  };

  React.useEffect(() => {
    fetchAllNamespaces();
  }, []);

  const renderContent = () => {
    if (!currentUser) {
      //不应该走这个逻辑
      return <Spin size="large" tip={intl.formatMessage({ id: 'pages.index.namespace.loadingUserInfo' })}/>;
    }

    if (viewMode === 'select') {
      return (
        <div style={{display: 'flex', gap: '24px', justifyContent: 'center'}}>
          <Card
            hoverable
            style={{
              width: 300,
              height: 150,
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center',
              boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)'
            }}
            onClick={() => setViewMode('create')}
          >
            <Typography.Title level={3}>
              <FormattedMessage
                id="pages.index.create.namespace"/>
            </Typography.Title>
          </Card>
          <Card
            hoverable
            style={{
              width: 300,
              height: 150,
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center',
              boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)'
            }}
            onClick={() => setViewMode('join')}
          >
            <Typography.Title level={3}>
              <FormattedMessage id="pages.index.select.namespace"/>
            </Typography.Title>
          </Card>
        </div>
      );
    }

    if (viewMode === 'create') {
      return (
        <Card
          title={<Typography.Title level={4}>
            <FormattedMessage
              id="pages.index.create.namespace"/>
          </Typography.Title>}
          style={{width: 640, boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)'}}
          extra={<Button type="link" onClick={() => setViewMode('select')}>
            <FormattedMessage id="pages.system.goBack"/>
          </Button>}
        >
          <p style={{marginBottom: 24}}>
            <FormattedMessage id="pages.index.namespace.create"/>
          </p>
          <CreateUpdateForm
            isCreate={true}
            key="create"
            run={(params: Partial<API.NamespaceCreateUpdate>) => {
              //创建空间并且刷新token
              return handleCreateNamespace(params)
            }}
          />

        </Card>
      );
    }

    if (viewMode === 'join') {
      return (
        <Card
          title={<Typography.Title level={4}>
            {(currentUser.namespaceMembers && currentUser.namespaceMembers.length > 0) ?
              <FormattedMessage id="pages.index.namespace.select"/> :
              <FormattedMessage id="pages.index.namespace.join"/>}
          </Typography.Title>}
          style={{width: 640, textAlign: 'center', boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)'}}
          extra={<Button type="link" onClick={() => setViewMode('select')}>
            <FormattedMessage id="pages.system.goBack"/>
          </Button>}
        >
          <p style={{marginBottom: 24}}>
            <FormattedMessage id='pages.index.namespace.selectAndJoin'/>
          </p>
          <Select
            showSearch
            style={{width: '100%'}}
            placeholder={intl.formatMessage({
              id: 'pages.index.namespace.searchSelect',
            })}
            optionFilterProp="children"
            filterOption={(input, option) =>
              (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
            }
            options={allNamespaces.map((ns) => {
              let statusLabel = '';
              switch (ns.status) {
                case 'APPROVED':
                  statusLabel = intl.formatMessage({ id: 'pages.index.namespace.statusApproved' });
                  break;
                case 'PENDING':
                  statusLabel = intl.formatMessage({ id: 'pages.index.namespace.statusPending' });
                  break;
                case 'REJECTED':
                  statusLabel = intl.formatMessage({ id: 'pages.index.namespace.statusRejected' });
                  break;
                default:
                  break;
              }
              //如果是自己的则展示出来
              const isMyNamespace = currentUser?.namespaceMembers?.some(item => item.id === ns.id);
              if (isMyNamespace) {
                statusLabel = intl.formatMessage({ id: 'pages.index.namespace.statusMine' });
              }
              return {
                label: intl.formatMessage({ id: 'pages.index.namespace.labelFormat' }, { name: ns.name, tag: ns.tag, status: statusLabel }),
                value: ns.id,
              };
            })}
            onSelect={(value) => {
              setSelectedNamespaceId(value);
            }}
            loading={loading}
            size="large"
          />
          <Button
            type="primary"
            size="large"
            style={{marginTop: 24, width: '100%'}}
            disabled={!selectedNamespaceId || allNamespaces.find(ns => ns.id === selectedNamespaceId)?.status === 'PENDING'}
            loading={loading}
            onClick={() => {
              const selectedNamespace = allNamespaces.find(ns => ns.id === selectedNamespaceId);
              if (!selectedNamespace) return;

              if(currentUser.is_super_admin){
                handleSelectNamespace(selectedNamespace);
                return;
              }

              switch (selectedNamespace.status) {
                //进入
                case 'APPROVED':
                  handleSelectNamespace(selectedNamespace);
                  break;
                case "":
                case undefined:
                case "REJECTED":
                  //重新申请,第一次申请
                  Modal.confirm({
                    title: intl.formatMessage({ id: 'pages.index.namespace.applyModalTitle' }),
                    content: intl.formatMessage({ id: 'pages.index.namespace.applyModalContent' }, { name: selectedNamespace.name }),
                    onOk: () => handleRequestToJoin(selectedNamespace),
                  });
                  break;
              }
            }}
          >
            {(() => {
              if (!selectedNamespaceId) {
                return <FormattedMessage id="pages.index.namespace.selectFirst"/>;
              }
              const selectedNamespace = allNamespaces.find(ns => ns.id === selectedNamespaceId);
              if (currentUser.is_super_admin){
                return <FormattedMessage id="pages.index.namespace.enter"/>;
              }
              switch (selectedNamespace?.status) {
                //批准进入
                case 'APPROVED':
                  return <FormattedMessage id="pages.index.namespace.enter"/>;
                //审批中拒绝重复提交
                case 'PENDING':
                  return <FormattedMessage id="pages.index.namespace.pending"/>;
                //已拒绝重复申请
                case 'REJECTED':
                  return <FormattedMessage id="pages.index.namespace.reapply"/>;
                default:
                  //没有单子返回状态为空则申请加入
                  return <FormattedMessage id="pages.index.namespace.apply"/>;
              }
            })()}
          </Button>
        </Card>
      );
    }

    return null;
  };

  return (
    <div className={styles.container}>

      <Lang/>
      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          flex: '1',
          padding: '32px 0',
        }}
      >
        {renderContent()}
      </div>
      <Footer/>
    </div>
  );
};

export default Login;
