// @ts-ignore
/* eslint-disable */

declare namespace API {
  type ApprovalDetailView = {
    data: ApprovalDetailViewItem[],
    total: number,
    success: boolean,
  };

  type ApprovalDetailViewItem<T = any> = Approval & {
    approvable: T;
  };

  type Approval = CommonModel & {
    approvableType: string;
    details: string;
    status: ApprovalStatus;
    requesterUser: number;
    approverUsers: string;
    approvalNotes: string;
    approvalTime: string | null;

    approverUserStr: string;
    requesterUserStr: string;
  };

  type ApprovalStatus = 'PENDING' | 'REJECTED' | 'APPROVED';

  type ApprovalForm ={
    id: number;
    notes: string;
    status: number;
  }

}


