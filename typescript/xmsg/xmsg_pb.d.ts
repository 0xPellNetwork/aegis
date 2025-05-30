// @generated by protoc-gen-es v1.3.0 with parameter "target=dts"
// @generated from file xmsg/xmsg.proto (package pellchain.pellcore.xmsg, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import type { InboundPellEvent } from "./pell_event_pb.js";

/**
 * @generated from enum pellchain.pellcore.xmsg.XmsgStatus
 */
export declare enum XmsgStatus {
  /**
   * some observer sees inbound tx
   *
   * @generated from enum value: PendingInbound = 0;
   */
  PendingInbound = 0,

  /**
   * super majority observer see inbound tx
   *
   * @generated from enum value: PendingOutbound = 1;
   */
  PendingOutbound = 1,

  /**
   * the corresponding outbound tx is mined
   *
   * @generated from enum value: OutboundMined = 2;
   */
  OutboundMined = 2,

  /**
   * outbound cannot succeed; should revert inbound
   *
   * @generated from enum value: PendingRevert = 3;
   */
  PendingRevert = 3,

  /**
   * inbound reverted.
   *
   * @generated from enum value: Reverted = 4;
   */
  Reverted = 4,

  /**
   * inbound tx error or invalid paramters and cannot revert; just abort.
   *
   * @generated from enum value: Aborted = 5;
   */
  Aborted = 5,
}

/**
 * @generated from enum pellchain.pellcore.xmsg.TxFinalizationStatus
 */
export declare enum TxFinalizationStatus {
  /**
   * the corresponding tx is not finalized
   *
   * @generated from enum value: NotFinalized = 0;
   */
  NotFinalized = 0,

  /**
   * the corresponding tx is finalized but not executed yet
   *
   * @generated from enum value: Finalized = 1;
   */
  Finalized = 1,

  /**
   * the corresponding tx is executed
   *
   * @generated from enum value: Executed = 2;
   */
  Executed = 2,
}

/**
 * inbound transaction param
 *
 * @generated from message pellchain.pellcore.xmsg.InboundTxParams
 */
export declare class InboundTxParams extends Message<InboundTxParams> {
  /**
   * this address is the immediate contract/EOA that calls
   *
   * @generated from field: string sender = 1;
   */
  sender: string;

  /**
   * the Connector.send()
   *
   * @generated from field: int64 sender_chain_id = 2;
   */
  senderChainId: bigint;

  /**
   * this address is the EOA that signs the inbound tx
   *
   * @generated from field: string tx_origin = 3;
   */
  txOrigin: string;

  /**
   * TODO: inbound_pell_event
   *
   * @generated from field: pellchain.pellcore.xmsg.InboundPellEvent inbound_pell_tx = 4;
   */
  inboundPellTx?: InboundPellEvent;

  /**
   * @generated from field: string inbound_tx_hash = 5;
   */
  inboundTxHash: string;

  /**
   * @generated from field: uint64 inbound_tx_block_height = 6;
   */
  inboundTxBlockHeight: bigint;

  /**
   * @generated from field: uint64 inbound_tx_event_index = 7;
   */
  inboundTxEventIndex: bigint;

  /**
   * @generated from field: string inbound_tx_ballot_index = 8;
   */
  inboundTxBallotIndex: string;

  /**
   * @generated from field: uint64 inbound_tx_finalized_pell_height = 9;
   */
  inboundTxFinalizedPellHeight: bigint;

  /**
   * @generated from field: pellchain.pellcore.xmsg.TxFinalizationStatus tx_finalization_status = 10;
   */
  txFinalizationStatus: TxFinalizationStatus;

  constructor(data?: PartialMessage<InboundTxParams>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "pellchain.pellcore.xmsg.InboundTxParams";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): InboundTxParams;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): InboundTxParams;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): InboundTxParams;

  static equals(a: InboundTxParams | PlainMessage<InboundTxParams> | undefined, b: InboundTxParams | PlainMessage<InboundTxParams> | undefined): boolean;
}

/**
 * outbound transaction param
 *
 * @generated from message pellchain.pellcore.xmsg.OutboundTxParams
 */
export declare class OutboundTxParams extends Message<OutboundTxParams> {
  /**
   * @generated from field: string receiver = 1;
   */
  receiver: string;

  /**
   * @generated from field: int64 receiver_chain_id = 2;
   */
  receiverChainId: bigint;

  /**
   * @generated from field: uint64 outbound_tx_tss_nonce = 3;
   */
  outboundTxTssNonce: bigint;

  /**
   * @generated from field: uint64 outbound_tx_gas_limit = 4;
   */
  outboundTxGasLimit: bigint;

  /**
   * @generated from field: string outbound_tx_gas_price = 5;
   */
  outboundTxGasPrice: string;

  /**
   * @generated from field: string outbound_tx_gas_priority_fee = 6;
   */
  outboundTxGasPriorityFee: string;

  /**
   * the above are commands for pellclients
   * the following fields are used when the outbound tx is mined
   *
   * @generated from field: string outbound_tx_hash = 7;
   */
  outboundTxHash: string;

  /**
   * @generated from field: string outbound_tx_ballot_index = 8;
   */
  outboundTxBallotIndex: string;

  /**
   * @generated from field: uint64 outbound_tx_external_height = 9;
   */
  outboundTxExternalHeight: bigint;

  /**
   * @generated from field: string tss_pubkey = 10;
   */
  tssPubkey: string;

  /**
   * @generated from field: pellchain.pellcore.xmsg.TxFinalizationStatus tx_finalization_status = 11;
   */
  txFinalizationStatus: TxFinalizationStatus;

  /**
   * @generated from field: uint64 outbound_tx_gas_used = 12;
   */
  outboundTxGasUsed: bigint;

  /**
   * @generated from field: string outbound_tx_effective_gas_price = 13;
   */
  outboundTxEffectiveGasPrice: string;

  /**
   * @generated from field: uint64 outbound_tx_effective_gas_limit = 14;
   */
  outboundTxEffectiveGasLimit: bigint;

  constructor(data?: PartialMessage<OutboundTxParams>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "pellchain.pellcore.xmsg.OutboundTxParams";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): OutboundTxParams;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): OutboundTxParams;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): OutboundTxParams;

  static equals(a: OutboundTxParams | PlainMessage<OutboundTxParams> | undefined, b: OutboundTxParams | PlainMessage<OutboundTxParams> | undefined): boolean;
}

/**
 * @generated from message pellchain.pellcore.xmsg.Status
 */
export declare class Status extends Message<Status> {
  /**
   * @generated from field: pellchain.pellcore.xmsg.XmsgStatus status = 1;
   */
  status: XmsgStatus;

  /**
   * @generated from field: string status_message = 2;
   */
  statusMessage: string;

  /**
   * @generated from field: int64 last_update_timestamp = 3;
   */
  lastUpdateTimestamp: bigint;

  constructor(data?: PartialMessage<Status>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "pellchain.pellcore.xmsg.Status";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Status;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Status;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Status;

  static equals(a: Status | PlainMessage<Status> | undefined, b: Status | PlainMessage<Status> | undefined): boolean;
}

/**
 * @generated from message pellchain.pellcore.xmsg.Xmsg
 */
export declare class Xmsg extends Message<Xmsg> {
  /**
   * @generated from field: string creator = 1;
   */
  creator: string;

  /**
   * @generated from field: string index = 2;
   */
  index: string;

  /**
   * @generated from field: pellchain.pellcore.xmsg.Status xmsg_status = 3;
   */
  xmsgStatus?: Status;

  /**
   * @generated from field: pellchain.pellcore.xmsg.InboundTxParams inbound_tx_params = 4;
   */
  inboundTxParams?: InboundTxParams;

  /**
   * @generated from field: repeated pellchain.pellcore.xmsg.OutboundTxParams outbound_tx_params = 5;
   */
  outboundTxParams: OutboundTxParams[];

  constructor(data?: PartialMessage<Xmsg>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "pellchain.pellcore.xmsg.Xmsg";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Xmsg;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Xmsg;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Xmsg;

  static equals(a: Xmsg | PlainMessage<Xmsg> | undefined, b: Xmsg | PlainMessage<Xmsg> | undefined): boolean;
}

