import type {Settings as LayoutSettings} from '@ant-design/pro-components';
import {PageLoading} from '@ant-design/pro-components';
import {App, Select} from 'antd';
import type {RequestConfig, RunTimeLayoutConfig} from '@umijs/max';
import {history, getIntl} from '@umijs/max';
import React, {useState, useEffect} from 'react';
import {AvatarDropdown, AvatarName, SelectLang} from '@/components';
import defaultSettings from '../config/defaultSettings';
import {errorConfig} from './requestErrorConfig';
import '@ant-design/v5-patch-for-react-19';
import {currentUser as fetchCurrentUser, refreshUserToken, selectNamespaceLike} from '@/services/index/api';
import {loginPath, selectPath} from "@/path";
import {refreshToken} from "@/pages/user/useUser";
import { useErrorHandler, handleErrorWithoutHook } from '@/utils/useErrorHandler';

declare const API_URL: string;

/**
 * @see https://umijs.org/docs/api/runtime-config#getinitialstate
 * */
export async function getInitialState(): Promise<{
  settings?: Partial<LayoutSettings>;
  currentUser?: API.CurrentUser;
  loading?: boolean;
  fetchUserInfo?: () => Promise<API.CurrentUser | undefined>;
  refreshToken?: (params: API.RefreshToken) => Promise<API.LoginResult | undefined>;
  selectNamespace?: (params: API.SearchParams) => Promise<API.AllNamespaceWithUser[] | undefined>;
}> {
  const intl = getIntl();

  const fetchUserInfo = async () => {
    try {
      const response = await fetchCurrentUser();
      if (response) {
        return {
          ...response,
          avatar: 'https://gw.alipayobjects.com/zos/antfincdn/XAosXuNZyF/BiazfanxmamNRoxxVxka.png',
        } as unknown as API.CurrentUser;
      }
    } catch (error) {
      console.error('获取用户信息失败:', error);
      handleErrorWithoutHook(error, intl.formatMessage({ id: 'pages.index.fetchUser.fail' }))
      history.push(loginPath);
    }
    return undefined;
  };
  const selectNamespace = async (params: API.SearchParams) => {
    try {
      const response = await selectNamespaceLike(params);
      return response as unknown as API.AllNamespaceWithUser[];
    } catch (error) {
      console.error('获取列表失败:', error);
      handleErrorWithoutHook(error, intl.formatMessage({ id: 'pages.index.fetchNamespace.fail' }));
    }
    return undefined;
  };
  const refreshToken = async (params: API.RefreshToken) => {
    try {
      const response = await refreshUserToken(params);
      return response as unknown as API.LoginResult;
    } catch (error) {
      console.error('刷新token失败:', error);
      handleErrorWithoutHook(error, intl.formatMessage({ id: 'pages.index.refreshToken.fail' }));
    }
    return undefined;
  };

  const {location} = history;
  // 如果是登录页面，直接返回，不获取用户信息
  if ([loginPath].includes(location.pathname)) {
    return {fetchUserInfo,
      refreshToken,
      selectNamespace,
      settings: defaultSettings as Partial<LayoutSettings>};
  }

  // 其他页面，获取用户信息
  const currentUser = await fetchUserInfo();
  if (currentUser == undefined) {
    return {
      fetchUserInfo,
      refreshToken,
      selectNamespace,
      currentUser,
      settings: defaultSettings as Partial<LayoutSettings>,
    };
  }

  //此处是普通用户的处理逻辑(超级管理员同样处理)
  // 如果不是选择空间页面，也不是登录页面，要判断用户是否已选择
  if (![selectPath].includes(location.pathname)) {
    //合规超级管理员用户返回(用户属于超级管理员 & 用户选择了空间)
    if (currentUser?.is_super_admin && currentUser?.select_namespace!=="") {
      return {
        fetchUserInfo,
        refreshToken,
        selectNamespace,
        currentUser,
        settings: defaultSettings as Partial<LayoutSettings>,
      };
    }
    //合规用户返回(用户名下有空间 & 用户选择了空间)
    if (currentUser?.namespaceMembers && currentUser?.select_namespace!=="") {
      return {
        fetchUserInfo,
        refreshToken,
        selectNamespace,
        currentUser,
        settings: defaultSettings as Partial<LayoutSettings>,
      };
    }
    //不合规用户 跳选择页面
    history.push(selectPath);
    //此处的用户信息正常
    return {
      fetchUserInfo,
      refreshToken,
      selectNamespace,
      currentUser,
      settings: defaultSettings as Partial<LayoutSettings>,
    };
  }

  //选择页
  return {
    fetchUserInfo,
    refreshToken,
    selectNamespace,
    currentUser,
    settings: defaultSettings as Partial<LayoutSettings>,
  };
}

// ProLayout 支持的api https://procomponents.ant.design/components/layout
export const layout: RunTimeLayoutConfig = ({
                                              initialState,
                                              setInitialState,
                                            }) => {
  // antd 的 message 组件需要 App 包裹，我们在 layout 函数外部无法使用 useApp hook
  // 所以在这里获取 message 实例，传递给下面的事件处理函数
  const {message} = App.useApp();
  const [allNamespaces, setAllNamespaces] = useState<Select[]>([]);
  const [namespacesLoading, setNamespacesLoading] = useState(false);
  const {handleError} = useErrorHandler();

  // 获取所有命名空间（用于超级管理员）
  const fetchAllNamespaces = async () => {
    if (initialState?.currentUser?.is_super_admin) {

      try {
        setNamespacesLoading(true);
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
        console.error('获取所有命名空间失败:', error);
        handleError(error, "pages.index.fetchAllNamespace.fail", {showDetail: true})
      } finally {
        setNamespacesLoading(false);
      }
    }
  };

  // 初始加载
  useEffect(() => {
    fetchAllNamespaces();
  }, [initialState?.currentUser?.is_super_admin]);

  const handleNamespaceChange = async (newNamespaceTag: string) => {
    const intl = getIntl();
    try {
      const token = await initialState?.refreshToken?.({
        userId: initialState?.currentUser?.id,
        selectNamespace: newNamespaceTag,
      });
      if (token && token.token) {
        refreshToken(token.token)
        message.success(intl.formatMessage({ id: 'pages.app.namespace.switchSuccess' }));
        // 刷新页面
        window.location.reload();
      } else {
        message.error(intl.formatMessage({ id: 'pages.app.namespace.switchFailed' }));
      }
    } catch (e) {
      handleError(e, "pages.app.namespace.switchFailed", {showDetail: true})
    }
  };
  return {
    actionsRender: () => {
      const currentUser = initialState?.currentUser as any;
      const currentNamespace = currentUser?.select_namespace;

      let namespaceOptions = [];

      //超级管理员取的是全部
      const sourceNamespaces = currentUser?.is_super_admin
        ? allNamespaces
        : (currentUser?.namespaceMembers || []);

      const intl = getIntl();
      namespaceOptions = sourceNamespaces.map((ns: any) => {
        const namespace = currentUser?.is_super_admin ? ns : ns.namespace;
        return {
          value: namespace.tag,
          label: intl.formatMessage(
            { id: 'pages.app.namespace.currentSpace' },
            { name: namespace.name, tag: namespace.tag }
          ),
        };
      });

      return [
        <SelectLang key="SelectLang" />,
        <Select
          key="namespace-select"
          variant="borderless"
          style={{width: 370}}
          value={currentNamespace}
          onSelect={handleNamespaceChange}
          onOpenChange={(open) => {
            if (open && currentUser?.is_super_admin) {
              fetchAllNamespaces();
            }
          }}
          loading={namespacesLoading}
          options={namespaceOptions}
          placeholder={getIntl().formatMessage({ id: 'pages.app.namespace.selectPlaceholder' })}
        />,
      ];
    },
    avatarProps: {
      src: initialState?.currentUser?.avatar,
      title: <AvatarName/>,
      render: (_, avatarChildren) => (
        <AvatarDropdown>{avatarChildren}</AvatarDropdown>
      ),
    },
    waterMarkProps: {
      content: initialState?.currentUser?.username,
    },
    bgLayoutImgList: [
      {
        src: '/one.png',
        left: 85,
        bottom: 100,
        height: '303px',
      },
      {
        src: '/two.png',
        bottom: -68,
        right: -45,
        height: '303px',
      },
      {
        src: '/three.png',
        bottom: 0,
        left: 0,
        width: '331px',
      },
    ],
    menuHeaderRender: undefined,
    // 自定义 403 页面
    // unAccessible: <div>unAccessible</div>,
    // 增加一个 loading 的状态
    childrenRender: (children) => {
      if (initialState?.loading) return <PageLoading/>;
      return <>{children}</>;
    },
    ...initialState?.settings,
  };
};

/**
 * @name request 配置，可以配置错误处理
 * 它基于 axios 和 ahooks 的 useRequest 提供了一套统一的网络请求和错误处理方案。
 * @doc https://umijs.org/docs/max/request#配置
 */
export const request: RequestConfig = {
  baseURL: API_URL,
  ...errorConfig,
};
