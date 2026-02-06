import {ProForm, ProFormInstance, ProFormText, ProFormTextArea, StepsForm} from '@ant-design/pro-components';
import {Button, message, Modal} from 'antd';
import React, {useEffect, useRef, useState} from 'react';
import {useIntl, useRequest} from '@umijs/max';
import {createEnv, updateEnv} from '@/services/env/api';
import {FormattedMessage} from '@@/exports';
import DriverConfiguration, {TabExchangeHandler} from './DriverConfiguration';
import {RuleObject} from "rc-field-form/es/interface";
import Driver = API.Driver;
import { useErrorHandler } from '@/utils/useErrorHandler';


export type CreateFormProps = {
  isCreate: boolean,
  values?: API.EnvItem;
  reload: () => void;
  trigger: React.ReactNode;
  selectNamespace?: string;
};

const CreateUpdateForm: React.FC<CreateFormProps> = (props) => {
  const {values, trigger, isCreate, selectNamespace} = props;
  const intl = useIntl();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [current, setCurrent] = useState(0);
  const formMapRef = useRef<ProFormInstance | null>(null);
  const tabExchangeHandlerRef = useRef<TabExchangeHandler>(null);
  const { handleError } = useErrorHandler();

  useEffect(() => {
    if (isModalOpen && tabExchangeHandlerRef.current) {
      tabExchangeHandlerRef.current.isCreate(isCreate);
    }
  }, [isModalOpen, isCreate, values]);


  const [messageApi, contextHolder] = message.useMessage();

  const {run, loading} = useRequest((params: Partial<API.EnvCreateUpdate>) => {
    if (selectNamespace && selectNamespace !== "") {
      params.select_namespace = selectNamespace
    }
    // 针对驱动类型 前端组装传递
    if (params.drivers){
      params.drivers = params.drivers.map(item=>({
          ...item,
        driverType: item.driverType + "_" + item.usage
      }));
    }
    return isCreate ? createEnv(params) : updateEnv(params);
  }, {
    manual: true,
    onSuccess: () => {
      const successMsg = isCreate
        ? intl.formatMessage({id: 'pages.env.form.createSuccess'})
        : intl.formatMessage({id: 'pages.env.form.updateSuccess'});
      console.log(successMsg);
      messageApi.success(successMsg);
      setIsModalOpen(false);
      setCurrent(0);
      props.reload();
    },
    onError: (error) => {
      handleError(
        error,
        isCreate
          ? intl.formatMessage({id: 'pages.env.form.createError'})
          : intl.formatMessage({id: 'pages.env.form.updateError'}),
        { showDetail: true }
      );
      setIsModalOpen(false);
      setCurrent(0);
    },
  });

  const renderTrigger = () => {
    return <span onClick={() => setIsModalOpen(true)}>{trigger}</span>;
  };

  return (
    <>
      {contextHolder}
      {renderTrigger()}
      <StepsForm
        current={current}
        onCurrentChange={setCurrent}
        formRef={formMapRef}
        submitter={{
          render: (props, dom) => {
            if (props.step === 0) {
              const nextButton = dom[0];
              return [
                <Button key="reset" onClick={() => {
                  const form = props.form;
                  if (!form) {
                    return;
                  }
                  if (isCreate) {
                    form.resetFields();
                  } else {
                    form.setFieldsValue({name: undefined, description: undefined});
                  }
                }}>
                  <FormattedMessage
                    id="pages.system.clearAll"/>
                </Button>,
                <Button key="next" type="primary" onClick={() => nextButton.props.onClick()}>
                  <FormattedMessage
                    id="pages.system.next"/>
                </Button>,
              ];
            }

            if (props.step === 1) {
              const preButton = dom[0];
              const nextButton = dom[1];
              return [
                <Button key="pre" onClick={() => preButton.props.onClick()}>
                  <FormattedMessage
                    id="pages.system.pre"/>
                </Button>,
                <Button key="submit" type="primary" onClick={() => {
                  console.log(props.form)
                  nextButton.props.onClick()
                }}>
                  <FormattedMessage
                    id="pages.system.submit"/>
                </Button>,
              ];
            }
            return dom;
          },
        }}
        stepsFormRender={(dom, submitter) => {
          return (
            <Modal
              width={640}
              styles={{
                body: {
                  padding: '32px 40px 0px',
                },
              }}
              destroyOnHidden
              title={isCreate ? intl.formatMessage({
                id: 'pages.env.searchTable.createChild',
              }) : intl.formatMessage({
                id: 'pages.env.searchTable.editChild',
              })}
              open={isModalOpen}
              footer={submitter}
              onCancel={() => {
                setIsModalOpen(false);
                setCurrent(0);
              }}
            >
              {dom}
            </Modal>
          );
        }}
        onFinish={async (formValues) => {
          console.log('打印一下', formValues)
          let finalValues = formValues;

          //需要把drivers中的负数ID全部干掉
          const newDrivers = formValues.drivers.map((d: API.Driver) => {
            if (d.id < 0) {
              return {
                ...d,
                id: 0,
              }
            }
            return {
              ...d,
            }
          });
          finalValues = {
            ...values,
            ...formValues,
            drivers: newDrivers,
          };
          console.log('打印一下处理后的值', finalValues);
          await run(finalValues);
        }}
        formProps={{
          initialValues: values,
        }}
      >
        <StepsForm.StepForm
          title={isCreate ? intl.formatMessage({
            id: 'pages.env.searchTable.createChild',
          }) : intl.formatMessage({
            id: 'pages.env.searchTable.editChild',
          })}
        >
          <ProFormText
            name="name"
            label={intl.formatMessage({id: 'pages.env.form.name'})}
            fieldProps={{
              style: {
                width: 510
              }
            }}
            formItemProps={{
              style: {
                marginTop: "20px"
              },
              labelCol: {
                style: {
                  paddingBottom: "10px",
                }
              },
            }}
            rules={[{required: true, message: intl.formatMessage({id: 'pages.env.form.nameRequired'})}]}
          />
          <ProFormText
            name="tag"
            label={
              <>
                {intl.formatMessage({id: 'pages.env.form.tag'})}
                <span style={{color: '#999', fontSize: '12px', marginLeft: '8px'}}>
                {intl.formatMessage({id: 'pages.env.form.tagReadonly'})}
                </span>
              </>
            }
            fieldProps={{
              style: {
                width: 510,
              },
              disabled: !isCreate
            }}
            formItemProps={{
              style: {
                marginTop: "20px"
              },
              labelCol: {
                style: {
                  paddingBottom: "10px",
                }
              },
            }}
            rules={[{required: true, message: intl.formatMessage({id: 'pages.env.form.tagRequired'})}]}
          />
          <ProFormText
            name="publish_order"
            label={intl.formatMessage({id: 'pages.env.form.publishOrder'})}
            transform={(value) => Number(value)}
            fieldProps={{
              type: 'number',
              style: {
                width: 510,
              },
            }}
            formItemProps={{
              style: {
                marginTop: "20px"
              },
              labelCol: {
                style: {
                  paddingBottom: "10px",
                }
              },
            }}
            rules={[
              {required: true, message: intl.formatMessage({id: 'pages.env.form.publishOrderRequired'})},
              {pattern: /^\d+$/, message: intl.formatMessage({id: 'pages.env.form.publishOrderInvalid'})}
            ]}
          />
          <ProFormTextArea
            name="description"
            label={intl.formatMessage({id: 'pages.env.form.description'})}
            fieldProps={{
              style: {
                width: 510
              }
            }}
            formItemProps={{
              style: {
                marginTop: "20px",
                marginBottom: "60px"
              },
              labelCol: {
                style: {
                  paddingBottom: "10px",
                }
              },
            }}
            rules={[{required: true, message: intl.formatMessage({id: 'pages.env.form.descriptionRequired'})}]}
          />
        </StepsForm.StepForm>
        <StepsForm.StepForm
          title={isCreate ? intl.formatMessage({
            id: 'pages.env.searchTable.driver.createChild',
          }) : intl.formatMessage({
            id: 'pages.env.searchTable.driver.updateChild',
          })}
        >
          <ProForm.Item
            name="drivers"
            validateTrigger={[]}
            rules={[
              {
                validator: async (_: RuleObject, value: Driver[]) => {
                  const drivers = value || [];
                  let hasProducer : boolean;
                  let hasConsumer: boolean;

                  hasProducer = drivers.some(driver => driver.usage === 'producer');
                  hasConsumer = drivers.some(driver => driver.usage === 'consumer');

                  if (!tabExchangeHandlerRef.current) {
                    console.error("获取不到 tabExchangeHandlerRef.current")
                    return Promise.resolve();
                  }
                  let tab = tabExchangeHandlerRef.current.getActiveTab();

                  //判断当前所在的tab页，如果是producer & 没有producers那么弹出提示
                  //切换tab取消提示
                  if (tab === 'producer') {
                    if (!hasProducer) {
                      return Promise.reject(new Error(intl.formatMessage({id: 'pages.env.form.producerRequired'})));
                    }
                    if (!hasConsumer) {
                      tabExchangeHandlerRef.current.setActiveTab('consumer');
                      return Promise.reject(new Error(intl.formatMessage({id: 'pages.env.form.consumerRequired'})));
                    }
                  }

                  if (tab === 'consumer') {
                    if (!hasConsumer) {
                      return Promise.reject(new Error(intl.formatMessage({id: 'pages.env.form.consumerRequired'})));
                    }

                    if (!hasProducer) {
                      tabExchangeHandlerRef.current.setActiveTab('producer');
                      return Promise.reject(new Error(intl.formatMessage({id: 'pages.env.form.producerRequired'})));
                    }
                  }
                  return Promise.resolve();
                },
              },
            ]}>
            <DriverConfiguration ref={tabExchangeHandlerRef} onTabChange={() => {
              formMapRef.current?.setFields([{name: 'drivers', errors: []}]);
            }} onSaveClick={() => {
              formMapRef.current?.setFields([{name: 'drivers', errors: []}]);
            }}/>
          </ProForm.Item>
        </StepsForm.StepForm>
      </StepsForm>
    </>
  );
};

export default CreateUpdateForm;
