// @generated by protoc-gen-es v1.3.0 with parameter "target=dts"
// @generated from file lightclient/chain_state.proto (package pellchain.pellcore.lightclient, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";

/**
 * ChainState defines the overall state of the block headers for a given chain
 *
 * @generated from message pellchain.pellcore.lightclient.ChainState
 */
export declare class ChainState extends Message<ChainState> {
  /**
   * @generated from field: int64 chain_id = 1;
   */
  chainId: bigint;

  /**
   * @generated from field: int64 latest_height = 2;
   */
  latestHeight: bigint;

  /**
   * @generated from field: int64 earliest_height = 3;
   */
  earliestHeight: bigint;

  /**
   * @generated from field: bytes latest_block_hash = 4;
   */
  latestBlockHash: Uint8Array;

  constructor(data?: PartialMessage<ChainState>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "pellchain.pellcore.lightclient.ChainState";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ChainState;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ChainState;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ChainState;

  static equals(a: ChainState | PlainMessage<ChainState> | undefined, b: ChainState | PlainMessage<ChainState> | undefined): boolean;
}

