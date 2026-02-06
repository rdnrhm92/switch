import {PageContainer, ProFormText, ProFormTextArea} from '@ant-design/pro-components';
import {Form, message, Card, Space, Button, Alert, Switch} from 'antd';
import React, {useState, useEffect} from 'react';
import {createSwitch, updateSwitch, getSwitchDetails} from '@/services/switch/api';
import {useIntl, useRequest, history, useParams, useLocation, Access, useModel} from '@umijs/max';
import {ArrowLeftOutlined} from '@ant-design/icons';
import RuleBuilder from '../components/RuleBuilder';
import ApproverConfig from '../components/ApproverConfig';
import { validateParsedRule, formatValidationErrors, ValidationError } from '@/utils/ruleValidator';
import { switchFactorsLike } from '@/services/switch-factor/api';
import { processApprovalsForForm, parseApproverUsers } from '@/utils/approvalUtils';
import {useErrorHandler} from "@/utils/useErrorHandler";
import {useAccess} from "@@/exports";

const CreateUpdatePage: React.FC = () => {
  const intl = useIntl();
  const params = useParams();
  const location = useLocation();
  const [form] = Form.useForm();
  const [messageApi, contextHolder] = message.useMessage();
  const { handleError } = useErrorHandler();
  const {initialState} = useModel('@@initialState');
  const {currentUser} = initialState || {};

  // 判断是创建还是编辑
  const isCreate = !params.id;
  const switchId = params.id ? parseInt(params.id as string) : undefined;

  const [initialValues, setInitialValues] = useState<API.SwitchModel | undefined>();
  const [validationErrors, setValidationErrors] = useState<ValidationError[]>([]);
  const [factorOptions, setFactorOptions] = useState<any[]>([]);
  const access = useAccess() as Record<string, boolean>;

  // 获取因子选项数据
  useEffect(() => {
    const fetchFactorOptions = async () => {
      try {
        const response = await switchFactorsLike({}, {}, {});
        setFactorOptions(response || []);
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

  // 编辑模式下渲染开关(基本信息+规则信息+审核信息)
  useEffect(() => {
    if (!isCreate && switchId) {
      const fetchSwitchDetail = async () => {
        try {
          const result = await getSwitchDetails(switchId);
          const switchData = result.data.factor;
          setInitialValues(switchData);

          const processedApprovals = processApprovalsForForm(result.data.approvals);

          // 设置表单字段值
          form.setFieldsValue({
            name: switchData.name,
            description: switchData.description,
            useCache: switchData.useCache,
            ruleConfig: switchData.rules,
            createSwitchApproversReq: processedApprovals
          });
        } catch (error) {
          handleError(
            error,
            'pages.switch.message.fetchSwitchDetailFailed',
            { showDetail: true }
          );
        }
      };
      fetchSwitchDetail();
    }
  }, [isCreate, switchId]);

  const {run, loading} = useRequest(
    (params: API.CreateUpdateSwitchReq) => {
      return isCreate ? createSwitch(params) : updateSwitch(params);
    },
    {
      manual: true,
      onSuccess: () => {
        messageApi.success(intl.formatMessage({ id: isCreate ? 'pages.switch.message.createSuccess' : 'pages.switch.message.updateSuccess' }));
        // 返回列表页
        history.push('/list/switch');
      },
      onError: (error) => {
        handleError(
          error,
          isCreate ? 'pages.switch.message.createFailed' : 'pages.switch.message.updateFailed',
          { showDetail: true }
        );
      },
    },
  );

  // 递归解析config字符串为JSON对象
  const parseConfigToJson = (node: any): any => {
    const parsedNode = { ...node };

    // 如果是因子节点，解析config
    if (parsedNode.config && typeof parsedNode.config === 'string') {
      try {
        parsedNode.config = JSON.parse(parsedNode.config);
      } catch (error) {
        // 如果解析失败，保持原值
        console.warn(intl.formatMessage({ id: 'pages.switch.message.configParseFailed' }) + ':', parsedNode.config, error);
      }
    }

    // 递归处理子节点
    if (parsedNode.children) {
      parsedNode.children = parsedNode.children.map((child: any) => parseConfigToJson(child));
    }

    return parsedNode;
  };

  const handleSubmit = async (formValues: any) => {
    // 清除之前的校验错误
    setValidationErrors([]);

    // 规则配置校验
    if (!formValues.ruleConfig) {
      messageApi.error(intl.formatMessage({ id: 'pages.switch.message.pleaseConfigureRule' }));
      return;
    }

    // JSON Schema 校验
    // 拷贝原始数据并解析value从string到json
    const parsedRule = parseConfigToJson(formValues.ruleConfig);
    console.log('解析后的真实JSON:', JSON.stringify(parsedRule, null, 2));

    const validationResult = validateParsedRule(parsedRule, factorOptions);

    if (!validationResult.isValid) {
      setValidationErrors(validationResult.errors);
      const errorMessage = formatValidationErrors(validationResult.errors, intl.formatMessage);
      messageApi.error(`${intl.formatMessage({ id: 'pages.switch.message.ruleValidationFailed' })}：\n${errorMessage}`);
      return;
    }

    // 审批人配置校验
    const approverConfigs = formValues.createSwitchApproversReq || [];
    for (let i = 0; i < approverConfigs.length; i++) {
      const config = approverConfigs[i];
      if (!config.envTag) {
        messageApi.error(intl.formatMessage({ id: 'pages.switch.message.approverEnvRequired' }, { index: i + 1 }));
        return;
      }

      // 审批人数组解析校验
      const approverUsers = parseApproverUsers(config.approverUsers);
      if (!approverUsers || approverUsers.length === 0) {
        messageApi.error(intl.formatMessage({ id: 'pages.switch.message.approverUsersRequired' }, { index: i + 1 }));
        return;
      }
    }

    const processedApproverConfigs = processApprovalsForForm(approverConfigs);

    const params = {
      ...initialValues,
      ...formValues,
      rules: parsedRule,
      // 后端接受的是uint数组。转换一下
      createSwitchApproversReq: processedApproverConfigs
    };
    if (!isCreate && switchId) {
      params.switchId = switchId;
    }
    await run(params);
  };

  const handleReset = () => {
    // 获取当前的开关名称
    const currentName = form.getFieldValue('name');

    form.resetFields();
    form.setFieldsValue({
      name: currentName,
      description: '',
      useCache: false,
      ruleConfig: null,
      createSwitchApproversReq: []
    });
  };

  const handleBack = () => {
    history.push('/list/switch');
  };

  // 判断是否应该禁用审批配置
  // 当不是创建模式，且创建者不是当前用户时，禁用
  // 因为主张谁创建了开关 谁有权定义审批人
  const isApproverConfigDisabled = !isCreate && initialValues?.createdBy !== currentUser?.username;

  return (
    <PageContainer
      header={{
        title: intl.formatMessage({ id: isCreate ? 'pages.switch.createUpdate.createTitle' : 'pages.switch.createUpdate.editTitle' }),
        breadcrumb: {
          items: [
            {
              path: '/list/switch',
              title: intl.formatMessage({ id: 'pages.switch.createUpdate.switchManagement' }),
            },
            {
              title: intl.formatMessage({ id: isCreate ? 'pages.switch.createUpdate.createTitle' : 'pages.switch.createUpdate.editTitle' }),
            },
          ],
        },
        extra: [
          <Button key="back" icon={<ArrowLeftOutlined />} onClick={handleBack}>
            {intl.formatMessage({ id: 'pages.switch.createUpdate.backToList' })}
          </Button>
        ],
      }}
    >
      {contextHolder}
      <Card>
        <Form
          form={form}
          layout="vertical"
          initialValues={initialValues}
          onFinish={handleSubmit}
          style={{ maxWidth: 1200 }}
        >
          <ProFormText
            name="name"
            label={
              <>
                {intl.formatMessage({ id: 'pages.switch.createUpdate.switchName' })}
                <span style={{color: '#999', fontSize: '12px', marginLeft: '8px'}}>
                  {intl.formatMessage({ id: 'pages.switch.createUpdate.switchNameReadonly' })}
                </span>
              </>
            }
            fieldProps={{
              disabled: !isCreate
            }}
            rules={[{required: true, message: intl.formatMessage({ id: 'pages.switch.createUpdate.switchNameRequired' })}]}
          />

          <ProFormTextArea
            name="description"
            label={intl.formatMessage({ id: 'pages.switch.createUpdate.switchDescription' })}
            rules={[{required: true, message: intl.formatMessage({ id: 'pages.switch.createUpdate.switchDescriptionRequired' })}]}
          />

          <Form.Item
            name="useCache"
            label={
              <>
                {intl.formatMessage({ id: 'pages.switch.createUpdate.enableCache' })}
                <span style={{
                  fontSize: '12px',
                  color: '#faad14',
                  marginLeft: '8px'
                }}>
                  {intl.formatMessage({ id: 'pages.switch.createUpdate.cacheWarning' })}
                </span>
              </>
            }
            valuePropName="checked"
            initialValue={false}
            style={{ marginTop: '20px' }}
          >
            <Switch
              checkedChildren={intl.formatMessage({ id: 'pages.switch.createUpdate.on' })}
              unCheckedChildren={intl.formatMessage({ id: 'pages.switch.createUpdate.off' })}
            />
          </Form.Item>

          <div style={{ marginTop: '20px', marginBottom: '20px' }}>
            <Form.Item label={intl.formatMessage({ id: 'pages.switch.createUpdate.ruleConfig' })} name="ruleConfig">
              <RuleBuilder />
            </Form.Item>
          </div>

          {/*{审批方面}*/}
          <Access accessible={!!access['users:list-for-switch']} key="create">
            <ApproverConfig disabled={isApproverConfigDisabled} />
          </Access>

          <Form.Item style={{ marginTop: '40px' }}>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                {intl.formatMessage({ id: isCreate ? 'pages.switch.createUpdate.create' : 'pages.switch.createUpdate.update' })}
              </Button>
              <Button onClick={handleReset}>
                {intl.formatMessage({ id: 'pages.switch.createUpdate.reset' })}
              </Button>
              <Button onClick={handleBack}>
                {intl.formatMessage({ id: 'pages.switch.createUpdate.cancel' })}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </PageContainer>
  );
};

export default CreateUpdatePage;
