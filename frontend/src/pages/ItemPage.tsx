import React, { ReactElement } from 'react';
import { ItemData } from '../dataObjects/item';
import { createStyles, Theme, Typography, WithStyles } from '@material-ui/core';
import Grid from '@material-ui/core/Grid';
import Divider from '@material-ui/core/Divider';
import Avatar from '@material-ui/core/Avatar';
import { Link as RouteLink, RouteComponentProps } from 'react-router-dom';
import { routes } from '../routes/Route';
import { StyleRules } from '@material-ui/core/styles';
import withStyles from '@material-ui/core/styles/withStyles';
import { ErrorProps, PageComponentWithError } from '../hoc/withBaseComponent';
import BasePageContainer from '../containers/BasePageContainer';
import LoadingComponent from '../components/LoadingComponent';
import { ItemFooter } from '../components/ItemFooter';
import { ItemImage } from '../components/ItemImage';

const styles = (theme: Theme): StyleRules =>
  createStyles({
    root: {
      marginBottom: theme.spacing(10),
    },
    title: {
      margin: theme.spacing(3),
    },
    avatar: {
      width: '50px',
      height: '50px',
    },
    divider: {
      margin: theme.spacing(1),
    },
    descSection: {
      marginTop: theme.spacing(3),
      marginBottom: theme.spacing(3),
    },
    link: {
      textDecoration: 'none',
    },
  });

interface ItemPageProps extends WithStyles<typeof styles> {
  loading: boolean;
  item: ItemData;
  viewer: {
    userId: number;
  };
  load: (itemId: string) => void;
  onClickBuy: (itemId: number) => void;
  onClickItemEdit: (itemId: number) => void;
  onClickBump: (itemId: number) => void;
  onClickTransaction: (itemId: number) => void;
}

type Props = ItemPageProps &
  RouteComponentProps<{ item_id: string }> &
  ErrorProps;

class ItemPage extends React.Component<Props> {
  constructor(props: Props) {
    super(props);

    this.props.load(this.props.match.params.item_id);
    this._onClickBuyButton = this._onClickBuyButton.bind(this);
    this._onClickItemEditButton = this._onClickItemEditButton.bind(this);
    this._onClickBumpButton = this._onClickBumpButton.bind(this);
    this._onClickTransaction = this._onClickTransaction.bind(this);
  }

  _onClickBuyButton(e: React.MouseEvent) {
    e.preventDefault();
    this.props.onClickBuy(this.props.item.id);
  }

  _onClickItemEditButton(e: React.MouseEvent) {
    e.preventDefault();
    this.props.onClickItemEdit(this.props.item.id);
  }

  _onClickBumpButton(e: React.MouseEvent) {
    e.preventDefault();
    this.props.onClickBump(this.props.item.id);
  }

  _onClickTransaction(e: React.MouseEvent) {
    e.preventDefault();
    this.props.onClickTransaction(this.props.item.id);
  }

  render() {
    const { classes, item, loading, viewer } = this.props;

    if (loading) {
      return <LoadingComponent />;
    }

    let buttons: {
      onClick: (e: React.MouseEvent) => void;
      buttonText: string;
      disabled: boolean;
      tooltip?: ReactElement;
    }[] = [
      {
        onClick: this._onClickBuyButton,
        buttonText: '??????',
        disabled: false,
      },
    ];

    // ?????????????????????????????????????????????????????? / Bump??????????????????
    if (viewer.userId === item.sellerId && item.status === 'on_sale') {
      buttons = [
        {
          onClick: this._onClickBumpButton,
          buttonText: 'Bump',
          disabled: false,
          tooltip: (
            <React.Fragment>
              <Typography variant="subtitle1">????????????</Typography>
              <Typography variant="subtitle2">
                BUMP??????????????????????????????????????????????????????????????????'
              </Typography>
            </React.Fragment>
          ),
        },
        {
          onClick: this._onClickItemEditButton,
          buttonText: '????????????',
          disabled: false,
        },
      ];
    }

    // ????????? or ??????????????????????????????????????????????????????????????????????????????
    if (
      (viewer.userId === item.sellerId || viewer.userId === item.buyerId) &&
      (item.status === 'trading' || item.status === 'sold_out')
    ) {
      buttons = [
        {
          onClick: this._onClickTransaction,
          buttonText: '????????????',
          disabled: false,
        },
      ];
    }

    // ???????????????????????????????????????????????????????????????????????????????????????
    if (
      item.status !== 'on_sale' &&
      viewer.userId !== item.sellerId &&
      viewer.userId !== item.buyerId
    ) {
      buttons = [
        {
          onClick(e: React.MouseEvent) {},
          buttonText: '????????????',
          disabled: true,
        },
      ];
    }

    return (
      <BasePageContainer>
        <div className={classes.root}>
          <Typography className={classes.title} variant="h3">
            {item.name}
          </Typography>
          <Grid container spacing={2}>
            <Grid item>
              <ItemImage
                imageUrl={item.thumbnailUrl}
                title={item.name}
                isSoldOut={item.status !== 'on_sale'}
                width={500}
              />
            </Grid>
            <Grid item xs={12} sm container>
              <Grid item xs container direction="column" spacing={2}>
                <Grid item xs>
                  <div className={classes.descSection}>
                    <Typography variant="h4">????????????</Typography>
                    <Divider className={classes.divider} variant="middle" />
                    <Typography variant="body1">{item.description}</Typography>
                  </div>

                  <div className={classes.descSection}>
                    <Typography variant="h4">????????????</Typography>
                    <Divider className={classes.divider} variant="middle" />
                    <Typography variant="body1">
                      <RouteLink
                        to={routes.categoryTimeline.getPath(
                          item.category.parentId,
                        )}
                      >
                        {item.category.parentCategoryName}
                      </RouteLink>{' '}
                      > {item.category.categoryName}
                    </Typography>
                  </div>

                  <div className={classes.descSection}>
                    <Typography variant="h4">?????????</Typography>
                    <Divider className={classes.divider} variant="middle" />
                    <Grid
                      container
                      direction="row"
                      justify="center"
                      alignItems="center"
                      wrap="nowrap"
                      spacing={2}
                    >
                      <Grid item>
                        <RouteLink
                          className={classes.link}
                          to={routes.user.getPath(item.sellerId)}
                        >
                          <Avatar className={classes.avatar}>
                            {item.seller.accountName.charAt(0)}
                          </Avatar>
                        </RouteLink>
                      </Grid>
                      <Grid item xs>
                        <Typography variant="body1">
                          {item.seller.accountName}
                        </Typography>
                      </Grid>
                    </Grid>
                  </div>
                </Grid>
              </Grid>
            </Grid>
          </Grid>
        </div>
        <ItemFooter price={item.price} buttons={buttons} />
      </BasePageContainer>
    );
  }
}

export default PageComponentWithError<any>()(withStyles(styles)(ItemPage));
