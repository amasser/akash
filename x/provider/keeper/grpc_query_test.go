package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/ovrclk/akash/app"
	"github.com/ovrclk/akash/testutil"
	"github.com/ovrclk/akash/x/provider/keeper"
	"github.com/ovrclk/akash/x/provider/types"
)

type grpcTestSuite struct {
	t      *testing.T
	app    *app.AkashApp
	ctx    sdk.Context
	keeper keeper.Keeper

	queryClient types.QueryClient
}

func setupTest(t *testing.T) *grpcTestSuite {
	suite := &grpcTestSuite{
		t: t,
	}

	key := sdk.NewKVStoreKey(types.StoreKey)

	suite.app = app.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{})
	suite.keeper = keeper.NewKeeper(types.ModuleCdc, key)
	querier := keeper.Querier{Keeper: suite.keeper}

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, querier)
	suite.queryClient = types.NewQueryClient(queryHelper)

	return suite
}

func TestGRPCQueryProvider(t *testing.T) {
	suite := setupTest(t)

	// creating provider
	provider := testutil.Provider(t)
	err := suite.keeper.Create(suite.ctx, provider)
	require.NoError(t, err)

	var (
		req         *types.QueryProviderRequest
		expProvider types.Provider
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = &types.QueryProviderRequest{}
			},
			false,
		},
		{
			"provider not found",
			func() {
				req = &types.QueryProviderRequest{Owner: testutil.AccAddress(t)}
			},
			false,
		},
		{
			"success",
			func() {
				req = &types.QueryProviderRequest{Owner: provider.Owner}
				expProvider = provider
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.msg), func(t *testing.T) {
			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)

			res, err := suite.queryClient.Provider(ctx, req)

			if tc.expPass {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Equal(t, expProvider, res.Provider)
			} else {
				require.Error(t, err)
				require.Nil(t, res)
			}

		})
	}
}

func TestGRPCQueryProviders(t *testing.T) {
	suite := setupTest(t)

	// creating providers
	provider := testutil.Provider(t)
	err := suite.keeper.Create(suite.ctx, provider)
	require.NoError(t, err)

	provider2 := testutil.Provider(t)
	err = suite.keeper.Create(suite.ctx, provider2)
	require.NoError(t, err)

	var (
		req         *types.QueryProvidersRequest
		expProvider types.Provider
	)

	testCases := []struct {
		msg      string
		malleate func()
		expLen   int
	}{
		{
			"query all providers without pagination",
			func() {
				req = &types.QueryProvidersRequest{}
				expProvider = provider
			},
			2,
		},
		{
			"query orders with pagination",
			func() {
				req = &types.QueryProvidersRequest{Pagination: &sdkquery.PageRequest{Limit: 1, Offset: 1}}
				expProvider = provider2
			},
			1,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.msg), func(t *testing.T) {
			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)

			res, err := suite.queryClient.Providers(ctx, req)

			require.NoError(t, err)
			require.NotNil(t, res)
			require.Equal(t, tc.expLen, len(res.Providers))
			require.Equal(t, expProvider, res.Providers[0])
		})
	}
}
