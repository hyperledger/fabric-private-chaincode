<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div>
    <AuctionHeader></AuctionHeader>
    <BidderMenu></BidderMenu>
    <BidSummaryTable></BidSummaryTable>
  </div>
</template>

<script>
import AuctionHeader from "../components/AppHeader";
import BidderMenu from "../components/AppMenu";
import BidSummaryTable from "../components/BidSummaryTable";

export default {
  name: "BidSummary",
  components: {
    AuctionHeader,
    BidderMenu,
    BidSummaryTable
  }
};
</script>
