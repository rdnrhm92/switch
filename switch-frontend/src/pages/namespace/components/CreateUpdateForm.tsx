import {ModalForm, ProFormText, ProFormTextArea} from '@ant-design/pro-components';
import {Button, Form, message} from 'antd';
import React, {useState} from 'react';
import {createNamespace, updateNamespace} from '@/services/namespace/api';
import {useIntl, useRequest} from '@umijs/max';
import {FormattedMessage} from "@@/plugin-locale";
import {useErrorHandler} from "@/utils/useErrorHandler";

export type CreateFormProps = {
  isCreate: boolean;
  values?: API.NamespaceItem;
  reload?: () => void;
  trigger?: React.ReactNode;
  run?: (params: Partial<API.NamespaceItem>) => Promise<any>;
  loading?: boolean;
};

const CreateUpdateForm: React.FC<CreateFormProps> = (props) => {
  const {values, trigger, isCreate, run: runFromProps, loading: loadingFromProps} = props;
  const intl = useIntl();
  // 如果 trigger 不存在，则默认打开 Modal
  const [isModalOpen, setIsModalOpen] = useState(!trigger);
  const [form] = Form.useForm();
  const { handleError } = useErrorHandler();

  const [messageApi, contextHolder] = message.useMessage();

  const {run: internalRun, loading: internalLoading} = useRequest(
    (params: Partial<API.NamespaceItem>) => {
      return isCreate ? createNamespace(params) : updateNamespace(params);
    },
    {
      manual: true,
      onSuccess: () => {
        const successMsg = intl.formatMessage({
          id: isCreate ? 'pages.namespace.form.createSuccess' : 'pages.namespace.form.updateSuccess'
        });
        console.log(successMsg);
        messageApi.success(successMsg);
        form.resetFields(); // 重置表单
        setIsModalOpen(false);
        props.reload && props.reload();
      },
      onError: (error) => {
        handleError(
          error,
          isCreate ? 'pages.namespace.form.createError' : 'pages.namespace.form.updateError',
          { showDetail: true }
        );
        setIsModalOpen(false);
      },
    },
  );

  const run = runFromProps || internalRun;
  const loading = loadingFromProps !== undefined ? loadingFromProps : internalLoading;

  const renderFormContent = () => (
    <div>
      <ProFormText
        name="name"
        label={intl.formatMessage({id: 'pages.namespace.form.name'})}
        formItemProps={{
          style: {
            marginTop: "15px",
          },
          labelCol: {
            style: {
              paddingBottom: "10px",
            }
          },
        }}
        fieldProps={{
          style: {
            width: "95%"
          },
        }}

        rules={[{required: true, message: intl.formatMessage({id: 'pages.namespace.form.nameRequired'})}]}
      />
      <ProFormText
        name="tag"
        label={
          <>
            {intl.formatMessage({id: 'pages.namespace.form.tag'})}
            <span style={{color: '#999', fontSize: '12px', marginLeft: '8px'}}>
                {intl.formatMessage({id: 'pages.namespace.form.tagReadonly'})}
                </span>
          </>
        }
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
        fieldProps={{
          style: {
            width: "95%"
          },
          disabled: !isCreate
        }}
        rules={[{required: true, message: intl.formatMessage({id: 'pages.namespace.form.tagRequired'})}]}
      />
      <ProFormTextArea
        name="description"
        label={intl.formatMessage({id: 'pages.namespace.form.description'})}
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
        fieldProps={{
          style: {
            width: "95%"
          }
        }}
        rules={[{required: true, message: intl.formatMessage({id: 'pages.namespace.form.descriptionRequired'})}]}
      />
    </div>
  );


  return (
    <>
      {contextHolder}
      {trigger ? (
        <div>
          <span onClick={() => setIsModalOpen(true)}>{trigger}</span>
          <ModalForm
            form={form}
            width={640}
            style={{
              paddingTop: '10px',
              marginLeft: "20px"
            }}
            labelCol={{flex: "0 0 25%"}}
            wrapperCol={{flex: "0 0 75%"}}
            title={isCreate ? intl.formatMessage({
              id: 'pages.namespace.searchTable.create'
            }) : intl.formatMessage({
              id: 'pages.namespace.searchTable.editChild',
            })}
            open={isModalOpen}
            onOpenChange={setIsModalOpen}
            initialValues={values}
            submitter={{
              searchConfig: {
                resetText: intl.formatMessage({id: 'pages.namespace.form.reset'}),
                submitText: intl.formatMessage({id: 'pages.namespace.form.submit'}),
              },
              submitButtonProps: {
                type: 'primary',
                style: {
                  marginLeft: "10px"
                },
                loading: loading,
              },
              resetButtonProps: {
                onClick: () => {
                  if (isCreate) {
                    form.resetFields();
                  } else {
                    form.setFieldsValue({name: undefined, description: undefined});
                  }
                },
              },
            }}
            modalProps={{
              styles: {
                header: {
                  paddingTop: '10px',
                  paddingLeft: '10px'
                },
              },
              destroyOnHidden: true
            }}
            onFinish={async (formValues) => {
              const params = {...values, ...formValues};
              await run(params);
            }}
          >
            {renderFormContent()}
          </ModalForm>
        </div>
      ) : (
        <Form
          form={form}
          layout="vertical"
          size="large"
          onFinish={async (formValues) => {
            const params = {...values, ...formValues};
            await run(params);
          }}
          initialValues={values}
        >
          {renderFormContent()}
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block>
              <FormattedMessage id="pages.index.namespace.createGo"/>
            </Button>
          </Form.Item>
        </Form>
      )}
    </>
  );
};

export default CreateUpdateForm;
