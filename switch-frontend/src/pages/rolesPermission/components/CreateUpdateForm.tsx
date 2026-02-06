import {ModalForm, ProFormCheckbox, ProFormText, ProFormTextArea} from '@ant-design/pro-components';
import {Button, Card, Col, Form, message, Modal, Row, Space, Typography, Input} from 'antd';
import {EditOutlined} from '@ant-design/icons';
import {useErrorHandler} from "@/utils/useErrorHandler";

const { TextArea } = Input;
import React, {useEffect, useState} from 'react';
import {
  allPermissions,
  insertRole,
  insertPermission,
  updatePermission,
  updateRole
} from '@/services/rolesPermission/api';
import {useIntl} from "@umijs/max";

const {Title, Text} = Typography;

export type FormValueType = Partial<API.RolesListItem>;

export type UpdateFormProps = {
  onCancel: () => void;
  onSubmit: (values: FormValueType) => Promise<void>;
  modalOpen: boolean;
  values: FormValueType;
};

const CreateUpdateForm: React.FC<UpdateFormProps> = (props) => {
  const [messageApi, contextHolder] = message.useMessage();
  const [form] = Form.useForm();
  const [addPermissionForm] = Form.useForm();
  const [editPermissionForm] = Form.useForm();
  const [permissions, setPermissions] = useState<API.RolesListPermissionItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [addPermissionModalOpen, setAddPermissionModalOpen] = useState(false);
  const [addingPermission, setAddingPermission] = useState(false);
  const [editPermissionModalOpen, setEditPermissionModalOpen] = useState(false);
  const [editingPermission, setEditingPermission] = useState<API.RolesListPermissionItem | null>(null);
  const [updatingPermission, setUpdatingPermission] = useState(false);
  const { handleError } = useErrorHandler();
  /**
   * @en-US International configuration
   * @zh-CN 国际化配置
   * */
  const intl = useIntl();

  const fetchPermissions = async () => {
    try {
      setLoading(true);
      const result = await allPermissions();
      setPermissions(result.data || []);
    } catch (error: any) {
      handleError(
        error,
        'pages.rolesPermission.message.permissionsLoadError',
        { showDetail: true }
      );
      setPermissions([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (props.modalOpen) {
      fetchPermissions();
    }
  }, [props.modalOpen]);

  useEffect(() => {
    if (permissions.length > 0 && props.modalOpen) {
      const permissionIds = props.values.permissions?.map(p => p.id) || [];
      form.setFieldsValue({...props.values, permissionIds});
    }
  }, [permissions, props.values, props.modalOpen, form]);

  const handleFinish = async (formValues: FormValueType) => {
    const params = {id: props.values.id, ...formValues};
    try {
      if (!params.id || params.id == 0){
        await insertRole(params);
      }else{
        await updateRole(params);
      }

      messageApi.success(intl.formatMessage({
        id: 'pages.rolesPermission.action.success',
      }));

      props.onSubmit(params);
    } catch (error) {
      handleError(
        error,
        'pages.rolesPermission.action.error',
        { showDetail: true }
      );
    }
  };

  const handleAddNewPermission = () => {
    setAddPermissionModalOpen(true);
  };

  const handleAddPermissionSubmit = async (values: API.UpsertPermission) => {
    try {
      setAddingPermission(true);
      await insertPermission(values);
      messageApi.success(intl.formatMessage({id: 'pages.rolesPermission.form.permissionAddSuccess'}));
      setAddPermissionModalOpen(false);
      addPermissionForm.resetFields();
      // 刷新权限列表
      await fetchPermissions();
    } catch (error: any) {
      handleError(
        error,
        'pages.rolesPermission.form.permissionAddError',
        { showDetail: true }
      );
    } finally {
      setAddingPermission(false);
    }
  };

  const handleEditPermission = (permission: API.RolesListPermissionItem) => {
    setEditingPermission(permission);
    setEditPermissionModalOpen(true);
    editPermissionForm.setFieldsValue({
      name: permission.name,
      description: permission.description,
    });
  };

  const handleUpdatePermissionSubmit = async (values: API.UpsertPermission) => {
    if (!editingPermission) return;

    try {
      setUpdatingPermission(true);
      await updatePermission({ ...values, id: editingPermission.id });
      messageApi.success(intl.formatMessage({id: 'pages.rolesPermission.form.permissionUpdateSuccess'}));
      setEditPermissionModalOpen(false);
      setEditingPermission(null);
      editPermissionForm.resetFields();
      // 刷新权限列表
      await fetchPermissions();
    } catch (error: any) {
      handleError(
        error,
        'pages.rolesPermission.form.permissionUpdateError',
        { showDetail: true }
      );
    } finally {
      setUpdatingPermission(false);
    }
  };

  return (
    <>
      {contextHolder}
      <ModalForm
        title={intl.formatMessage({
          id: props.values.id ? 'pages.rolesPermission.form.editRole' : 'pages.rolesPermission.form.createRole'
        })}
        width="640px"
        form={form}
        open={props.modalOpen}
        onFinish={handleFinish}
        onOpenChange={(visible) => {
          if (!visible) {
            form.resetFields();
            props.onCancel();
          }
        }}
        modalProps={{
          destroyOnHidden: true,
        }}
      >
      <Row gutter={24}>
        <Col span={10}>
          <Card
            title={<Title level={5} style={{margin: 0}}>{intl.formatMessage({id: 'pages.rolesPermission.form.roleInfo'})}</Title>}
            size="small"
            style={{height: '100%'}}
          >
            <Space direction="vertical" style={{width: '100%'}} size="large">
              <ProFormText
                name="name"
                label={intl.formatMessage({id: 'pages.rolesPermission.form.roleName'})}
                rules={[{required: true, message: intl.formatMessage({id: 'pages.rolesPermission.form.roleNameRequired'})}]}
              />
              <ProFormTextArea
                name="description"
                label={intl.formatMessage({id: 'pages.rolesPermission.form.roleDescription'})}
                placeholder={intl.formatMessage({id: 'pages.rolesPermission.form.roleDescriptionPlaceholder'})}
                rows={4}
              />
            </Space>
          </Card>
        </Col>

        <Col span={14}>
          <Card
            title={
              <Space>
                <Title level={5} style={{margin: 0}}>{intl.formatMessage({id: 'pages.rolesPermission.form.permissionAssignment'})}</Title>
                <Text type="secondary">({permissions.length} {intl.formatMessage({id: 'pages.rolesPermission.form.permissionsAvailable'})})</Text>
              </Space>
            }
            size="small"
            style={{height: '100%'}}
            extra={
              <Button type="link" onClick={handleAddNewPermission} size="small">
                {intl.formatMessage({id: 'pages.rolesPermission.form.addNewPermission'})}
              </Button>
            }
          >
            {loading ? (
              <div style={{textAlign: 'center', padding: '40px 0', color: '#999'}}>
                {intl.formatMessage({id: 'pages.rolesPermission.form.loadingPermissions'})}
              </div>
            ) : (
              <div style={{maxHeight: '300px', overflowY: 'auto'}}>
                <style>
                  {`
                    .permission-item:hover .permission-actions {
                      opacity: 1 !important;
                    }
                  `}
                </style>
                <ProFormCheckbox.Group
                  name="permissionIds"
                  options={permissions.map((p: API.RolesListPermissionItem) => ({
                    label: (
                      <div className="permission-item" style={{display: 'flex', justifyContent: 'space-between', alignItems: 'center', width: '100%'}}>
                        <Space direction="vertical" size={0}>
                          <Text style={{color: '#595959', fontWeight: 500}}>{p.name}</Text>
                          {p.description && <Text type="secondary" style={{fontSize: '12px'}}>{p.description}</Text>}
                        </Space>
                        {p.namespaceTag != "" && <Space size="small" style={{opacity: 0.3}} className="permission-actions">
                          <Button
                            type="text"
                            size="small"
                            icon={<EditOutlined />}
                            onClick={(e) => {
                              e.stopPropagation();
                              handleEditPermission(p);
                            }}
                            style={{color: '#1890ff'}}
                          />
                        </Space>}
                      </div>
                    ),
                    value: p.id
                  }))}
                />
              </div>
            )}
          </Card>
        </Col>
      </Row>

      <Modal
        title={intl.formatMessage({id: 'pages.rolesPermission.form.addPermission'})}
        open={addPermissionModalOpen}
        onCancel={() => {
          setAddPermissionModalOpen(false);
          addPermissionForm.resetFields();
        }}
        onOk={() => {
          addPermissionForm.validateFields().then(handleAddPermissionSubmit);
        }}
        confirmLoading={addingPermission}
      >
        <Form
          form={addPermissionForm}
          layout="vertical"
          style={{marginTop: '20px'}}
        >
          <Form.Item
            name="name"
            label={intl.formatMessage({id: 'pages.rolesPermission.form.permissionName'})}
            rules={[{required: true, message: intl.formatMessage({id: 'pages.rolesPermission.form.permissionNameRequired'})}]}
          >
            <Input
              placeholder={intl.formatMessage({id: 'pages.rolesPermission.form.permissionNamePlaceholder'})}
            />
          </Form.Item>
          <Form.Item
            name="description"
            label={intl.formatMessage({id: 'pages.rolesPermission.form.permissionDescription'})}
            rules={[{required: true, message: intl.formatMessage({id: 'pages.rolesPermission.form.permissionDescriptionRequired'})}]}
          >
            <TextArea
              placeholder={intl.formatMessage({id: 'pages.rolesPermission.form.permissionDescriptionPlaceholder'})}
              rows={3}
            />
          </Form.Item>
        </Form>
      </Modal>

      {/* 编辑权限弹框 */}
      <Modal
        title={intl.formatMessage({id: 'pages.rolesPermission.form.editPermission'})}
        open={editPermissionModalOpen}
        onCancel={() => {
          setEditPermissionModalOpen(false);
          setEditingPermission(null);
          editPermissionForm.resetFields();
        }}
        onOk={() => {
          editPermissionForm.validateFields().then(handleUpdatePermissionSubmit);
        }}
        confirmLoading={updatingPermission}
      >
        <Form
          form={editPermissionForm}
          layout="vertical"
          style={{marginTop: '20px'}}
        >
          <Form.Item
            name="name"
            label={intl.formatMessage({id: 'pages.rolesPermission.form.permissionName'})}
            rules={[{required: true, message: intl.formatMessage({id: 'pages.rolesPermission.form.permissionNameRequired'})}]}
          >
            <Input
              placeholder={intl.formatMessage({id: 'pages.rolesPermission.form.permissionNamePlaceholder'})}
            />
          </Form.Item>
          <Form.Item
            name="description"
            label={intl.formatMessage({id: 'pages.rolesPermission.form.permissionDescription'})}
            rules={[{required: true, message: intl.formatMessage({id: 'pages.rolesPermission.form.permissionDescriptionRequired'})}]}
          >
            <TextArea
              placeholder={intl.formatMessage({id: 'pages.rolesPermission.form.permissionDescriptionPlaceholder'})}
              rows={3}
            />
          </Form.Item>
        </Form>
      </Modal>
    </ModalForm>
    </>
  );
};

export default CreateUpdateForm;
