import {GithubOutlined} from '@ant-design/icons';
import {DefaultFooter} from '@ant-design/pro-components';
import React from 'react';

const Footer: React.FC = () => {
  return (
    <DefaultFooter
      style={{
        background: 'none',
      }}
      copyright="Powered by Switch"
      links={[
        {
          key: 'github',
          title: <GithubOutlined/>,
          href: 'https://gitee.com/fatzeng/collections/413616',
          blankTarget: true,
        },
        {
          key: 'gitee',
          title: <GithubOutlined/>,
          href: 'https://gitee.com/fatzeng/collections/413616',
          blankTarget: true,
        },
      ]}
    />
  );
};

export default Footer;
