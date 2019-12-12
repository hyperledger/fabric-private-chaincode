<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div>
    <AuctionHeader></AuctionHeader>
    <BidderMenu></BidderMenu>
    <MyResultsTable></MyResultsTable>
  </div>
</template>

<script>
import AuctionHeader from "../components/AppHeader";
import BidderMenu from "../components/AppMenu";
import MyResultsTable from "../components/MyResultsTable";

export default {
  name: "MyResults",
  components: {
    AuctionHeader,
    BidderMenu,
    MyResultsTable
  }
};
</script>
