// @generated by protoc-gen-es v1.3.0 with parameter "target=dts"
// @generated from file authority/tx.proto (package pellchain.pellcore.authority, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import type { Policies } from "./policies_pb.js";

/**
 * MsgUpdatePolicies defines the MsgUpdatePolicies service.
 *
 * @generated from message pellchain.pellcore.authority.MsgUpdatePolicies
 */
export declare class MsgUpdatePolicies extends Message<MsgUpdatePolicies> {
  /**
   * @generated from field: string signer = 1;
   */
  signer: string;

  /**
   * @generated from field: pellchain.pellcore.authority.Policies policies = 2;
   */
  policies?: Policies;

  constructor(data?: PartialMessage<MsgUpdatePolicies>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "pellchain.pellcore.authority.MsgUpdatePolicies";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): MsgUpdatePolicies;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): MsgUpdatePolicies;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): MsgUpdatePolicies;

  static equals(a: MsgUpdatePolicies | PlainMessage<MsgUpdatePolicies> | undefined, b: MsgUpdatePolicies | PlainMessage<MsgUpdatePolicies> | undefined): boolean;
}

/**
 * MsgUpdatePoliciesResponse defines the MsgUpdatePoliciesResponse service.
 *
 * @generated from message pellchain.pellcore.authority.MsgUpdatePoliciesResponse
 */
export declare class MsgUpdatePoliciesResponse extends Message<MsgUpdatePoliciesResponse> {
  constructor(data?: PartialMessage<MsgUpdatePoliciesResponse>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "pellchain.pellcore.authority.MsgUpdatePoliciesResponse";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): MsgUpdatePoliciesResponse;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): MsgUpdatePoliciesResponse;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): MsgUpdatePoliciesResponse;

  static equals(a: MsgUpdatePoliciesResponse | PlainMessage<MsgUpdatePoliciesResponse> | undefined, b: MsgUpdatePoliciesResponse | PlainMessage<MsgUpdatePoliciesResponse> | undefined): boolean;
}

