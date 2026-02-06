import {LockOutlined, UserOutlined} from '@ant-design/icons';
import {LoginForm, ProFormCheckbox, ProFormText} from '@ant-design/pro-components';
import {FormattedMessage, Helmet, SelectLang, useIntl,} from '@umijs/max';
import {Alert, App, Form, Tabs} from 'antd';
import {createStyles} from 'antd-style';
import React, {useState} from 'react';
import {Footer} from '@/components';
import {login, register} from '@/services/index/api';
import Settings from '../../../../config/defaultSettings';
import indexStyles from './index.less';
import {ResponseStructure} from "@/requestErrorConfig";
import {refreshToken, useUserActions} from "@/pages/user/useUser";
import {useModel} from "@@/exports";
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
      backgroundImage:
        "url('https://mdn.alipayobjects.com/yuyan_qk0oxh/afts/img/V-_oS6r-i7wAAAAAAAAAAAAAFl94AQBr')",
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
  const [userLoginState, setUserLoginState] = useState<ResponseStructure<API.LoginResult>>();
  const [activeKey, setActiveKey] = useState<string>("account");
  const {styles} = useStyles();
  const {message} = App.useApp();
  const intl = useIntl();
  const [form] = Form.useForm();
  const {refreshUserInfo} = useUserActions();
  const { handleError } = useErrorHandler();


  const handleSubmit = async (values: API.LoginParams) => {
    try {
      //登录 & 注册
      const msg = await (activeKey == "account" ? login({...values}) : register({...values}))
      
      if (msg.code == 0) {
        const defaultLoginSuccessMessage = activeKey == "account" ?
          intl.formatMessage({
            id: 'pages.index.login.success',
          }) : intl.formatMessage({
            id: 'pages.index.register.success',
          });
        message.success(defaultLoginSuccessMessage);
        
        //设置token
        if (msg.data && msg.data.token) {
          //设置token
          refreshToken(msg.data.token)
          
          // 先跳转到首页，让 getInitialState 重新加载用户信息并决定最终跳转
          // getInitialState 会检查用户是否需要选择命名空间
          const urlParams = new URL(window.location.href).searchParams;
          const redirectUrl = urlParams.get('redirect') || '/';
          window.location.href = redirectUrl;
          return;
        } else {
          console.error("未能获取到Authorization");
          return
        }
      }
      // 如果失败去设置用户错误信息
      setUserLoginState(msg);
    } catch (error) {
      const errorMsgKey = activeKey == "account" ? 'pages.index.login.failure' : 'pages.index.register.failure';
      handleError(
        error,
        errorMsgKey,
        { showDetail: true }
      );
    }
  };

  const clearError = (kind: string) => {
    form.setFields([
      {
        name: kind,
        errors: [],
      },
    ]);
  }

  return (
    <div className={styles.container}>
      <Helmet>
        <title>
          {intl.formatMessage({ id: 'menu.login' })}
          {Settings.title && ` - ${Settings.title}`}
        </title>
      </Helmet>
      <Lang/>
      <div
        style={{
          flex: '1',
          padding: '32px 0',
        }}
      >
        <LoginForm
          form={form}
          submitter={{searchConfig: {submitText: activeKey == "account" ? intl.formatMessage({ id: 'pages.index.login.submitText' }) : intl.formatMessage({ id: 'pages.index.register.submitText' }),}}}
          contentStyle={{
            minWidth: 280,
            maxWidth: '75vw',
          }}
          logo={
            <img
              alt="logo"
              src="/switch.svg"
              style={{
                width: 84,
                height: 84,
                position: 'relative',
                left: '-20px',
                top: '-23px',
                paddingBottom: '15px',
                paddingTop: '30px',
              }}
            />
          }
          title="Switch"
          subTitle={intl.formatMessage({
            id: 'pages.index.layouts.userLayout.title',
          })}
          initialValues={{
            autoLogin: true,
          }}
          onFinish={async (values) => {
            await handleSubmit(values as API.LoginParams);
          }}
        >
          <div style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between"
          }}>
            <Tabs
              rootClassName={indexStyles.loginTabs}
              activeKey={activeKey}
              onChange={(key) => {
                setActiveKey(key);
                setUserLoginState(undefined);
                form.setFieldsValue({username: undefined, password: undefined});
                clearError("username");
                clearError("password");
              }}
              items={[
                {
                  key: 'account',
                  label: intl.formatMessage({
                    id: 'pages.index.login.tab',
                  })
                },
                {
                  key: 'register',
                  label: intl.formatMessage({
                    id: 'pages.index.register.tab',
                  })
                },
              ]}
            />
          </div>


          {userLoginState?.code != 0 && userLoginState?.code != undefined && (
            <LoginMessage
              content={intl.formatMessage({
                id: 'pages.index.accountLogin.errorMessage',
              })}
            />
          )}
          <>
            <ProFormText
              name="username"
              fieldProps={{
                size: 'large',
                prefix: <UserOutlined/>,
                onChange: () => {
                  clearError("username")
                },
              }}
              placeholder={activeKey == "account" ?
                intl.formatMessage({
                  id: 'pages.index.usernameLogin.placeholder',
                }) : intl.formatMessage({
                  id: 'pages.index.usernameRegister.placeholder',
                })}
              rules={[
                {
                  required: true,
                  message: (
                    <FormattedMessage
                      id="pages.index.username.required"
                    />
                  ),
                },
              ]}
            />
            <ProFormText.Password
              name="password"
              fieldProps={{
                size: 'large',
                prefix: <LockOutlined/>,
                onChange: () => {
                  clearError("password")
                }
              }}
              placeholder={activeKey == "account" ?
                intl.formatMessage({
                  id: 'pages.index.passwordLogin.placeholder',
                }) : intl.formatMessage({
                  id: 'pages.index.passwordRegister.placeholder',
                })}
              rules={[
                {
                  required: true,
                  message: (
                    <FormattedMessage
                      id="pages.index.password.required"
                    />
                  ),
                },
              ]}
            />
          </>
          <div
            style={{
              marginBottom: 24,
            }}
          >
            <ProFormCheckbox noStyle name="autoLogin">
              <FormattedMessage
                id="pages.index.rememberMe"
              />
            </ProFormCheckbox>
          </div>
        </LoginForm>
      </div>
      <Footer/>
    </div>
  );
};

export default Login;
