// @generated by protoc-gen-es v1.3.0 with parameter "target=dts"
// @generated from file relayer/tss.proto (package pellchain.pellcore.relayer, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";

/**
 * tss
 *
 * @generated from message pellchain.pellcore.relayer.TSS
 */
export declare class TSS extends Message<TSS> {
  /**
   * @generated from field: string tss_pubkey = 3;
   */
  tssPubkey: string;

  /**
   * @generated from field: repeated string tss_participant_list = 4;
   */
  tssParticipantList: string[];

  /**
   * @generated from field: repeated string operator_address_list = 5;
   */
  operatorAddressList: string[];

  /**
   * @generated from field: int64 finalized_pell_height = 6;
   */
  finalizedPellHeight: bigint;

  /**
   * @generated from field: int64 keygen_pell_height = 7;
   */
  keygenPellHeight: bigint;

  constructor(data?: PartialMessage<TSS>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "pellchain.pellcore.relayer.TSS";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TSS;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TSS;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TSS;

  static equals(a: TSS | PlainMessage<TSS> | undefined, b: TSS | PlainMessage<TSS> | undefined): boolean;
}

