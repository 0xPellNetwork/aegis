// @generated by protoc-gen-es v1.3.0 with parameter "target=dts"
// @generated from file relayer/genesis.proto (package pellchain.pellcore.relayer, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import type { Ballot } from "./ballot_pb.js";
import type { LastRelayerCount, RelayerSet } from "./relayer_pb.js";
import type { NodeAccount } from "./node_account_pb.js";
import type { CrosschainFlags } from "./crosschain_flags_pb.js";
import type { ChainParamsList, Params } from "./params_pb.js";
import type { Keygen } from "./keygen_pb.js";
import type { TSS } from "./tss_pb.js";
import type { TssFundMigratorInfo } from "./tss_funds_migrator_pb.js";
import type { Blame } from "./blame_pb.js";
import type { PendingNonces } from "./pending_nonces_pb.js";
import type { ChainNonces } from "./chain_nonces_pb.js";
import type { NonceToXmsg } from "./nonce_to_xmsg_pb.js";

/**
 * relayer genesis state
 *
 * @generated from message pellchain.pellcore.relayer.GenesisState
 */
export declare class GenesisState extends Message<GenesisState> {
  /**
   * @generated from field: repeated pellchain.pellcore.relayer.Ballot ballots = 1;
   */
  ballots: Ballot[];

  /**
   * @generated from field: pellchain.pellcore.relayer.RelayerSet observers = 2;
   */
  observers?: RelayerSet;

  /**
   * @generated from field: repeated pellchain.pellcore.relayer.NodeAccount node_account_list = 3;
   */
  nodeAccountList: NodeAccount[];

  /**
   * @generated from field: pellchain.pellcore.relayer.CrosschainFlags crosschain_flags = 4;
   */
  crosschainFlags?: CrosschainFlags;

  /**
   * @generated from field: pellchain.pellcore.relayer.Params params = 5;
   */
  params?: Params;

  /**
   * @generated from field: pellchain.pellcore.relayer.Keygen keygen = 6;
   */
  keygen?: Keygen;

  /**
   * @generated from field: pellchain.pellcore.relayer.LastRelayerCount last_observer_count = 7;
   */
  lastObserverCount?: LastRelayerCount;

  /**
   * @generated from field: pellchain.pellcore.relayer.ChainParamsList chain_params_list = 8;
   */
  chainParamsList?: ChainParamsList;

  /**
   * @generated from field: pellchain.pellcore.relayer.TSS tss = 9;
   */
  tss?: TSS;

  /**
   * @generated from field: repeated pellchain.pellcore.relayer.TSS tss_history = 10;
   */
  tssHistory: TSS[];

  /**
   * @generated from field: repeated pellchain.pellcore.relayer.TssFundMigratorInfo tss_fund_migrators = 11;
   */
  tssFundMigrators: TssFundMigratorInfo[];

  /**
   * @generated from field: repeated pellchain.pellcore.relayer.Blame blame_list = 12;
   */
  blameList: Blame[];

  /**
   * @generated from field: repeated pellchain.pellcore.relayer.PendingNonces pending_nonces = 13;
   */
  pendingNonces: PendingNonces[];

  /**
   * @generated from field: repeated pellchain.pellcore.relayer.ChainNonces chain_nonces = 14;
   */
  chainNonces: ChainNonces[];

  /**
   * @generated from field: repeated pellchain.pellcore.relayer.NonceToXmsg nonce_to_xmsg = 15;
   */
  nonceToXmsg: NonceToXmsg[];

  constructor(data?: PartialMessage<GenesisState>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "pellchain.pellcore.relayer.GenesisState";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenesisState;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenesisState;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenesisState;

  static equals(a: GenesisState | PlainMessage<GenesisState> | undefined, b: GenesisState | PlainMessage<GenesisState> | undefined): boolean;
}

