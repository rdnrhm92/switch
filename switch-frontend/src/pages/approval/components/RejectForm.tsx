import { ModalForm, ProFormTextArea } from '@ant-design/pro-components';
import { Form, message } from 'antd';
import React, { useState } from 'react';
import { useIntl, useRequest } from '@umijs/max';
import {approveRequest} from '@/services/approval/api';
import { useErrorHandler } from '@/utils/useErrorHandler';

export type RejectFormProps = {
  values: API.ApprovalDetailViewItem;
  reload: () => void;
  trigger: React.ReactNode;
};

const RejectForm: React.FC<RejectFormProps> = (props) => {
  const { values, trigger } = props;
  const intl = useIntl();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [form] = Form.useForm();
  const { handleError } = useErrorHandler();

  const [messageApi, contextHolder] = message.useMessage();

  const { run, loading } = useRequest(approveRequest, {
    manual: true,
    onSuccess: () => {
      messageApi.success(intl.formatMessage({ id: 'pages.approvals.action.success' }));
      setIsModalOpen(false);
      props.reload();
    },
    onError: (error) => {
      handleError(
        error,
        'pages.approvals.action.error',
        {
          showDetail: true,
          onError: () => {
            setIsModalOpen(false);
          }
        }
      );
    },
  });

  const renderTrigger = () => {
    return <span onClick={() => setIsModalOpen(true)}>{trigger}</span>;
  };

  return (
    <>
      {contextHolder}
      {renderTrigger()}
      <ModalForm
        form={form}
        width={640}
        title={intl.formatMessage({ id: 'pages.approvals.action.rejectTitle' })}
        open={isModalOpen}
        onOpenChange={setIsModalOpen}
        onFinish={async (formValues) => {
          await run({
            id: values.id,
            notes: formValues.reason,
            status: 0
          });
        }}
      >
        <ProFormTextArea
          name="reason"
          label={intl.formatMessage({ id: 'pages.approvals.action.rejectReason' })}
          placeholder={intl.formatMessage({ id: 'pages.approvals.action.rejectReasonPlaceholder' })}
          rules={[{ required: true, message: intl.formatMessage({ id: 'pages.approvals.action.rejectReasonRequired' }) }]}
        />
      </ModalForm>
    </>
  );
};

export default RejectForm;
